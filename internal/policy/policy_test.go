package policy

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"golang.org/x/time/rate"

	"github.com/glefebvre/stalkeer/internal/config"
	"github.com/glefebvre/stalkeer/internal/models"
)

type fakeChecker struct {
	active bool
}

func (f *fakeChecker) ActivePlayback(ctx context.Context) bool {
	return f.active
}

func throttleCfg(jellyfinEnabled bool, jellyfinAction string, throttleKbps int) *config.Config {
	return &config.Config{
		Jellyfin: config.JellyfinConfig{
			PlaybackCheckEnabled: jellyfinEnabled,
			PlaybackAction:       jellyfinAction,
		},
		Downloads: config.DownloadsConfig{
			ThrottleRateKbps: throttleKbps,
		},
	}
}

func windowsWithAction(action string) []models.BandwidthScheduleWindow {
	if action == ActionNone {
		return nil
	}
	return []models.BandwidthScheduleWindow{
		{DaysOfWeek: models.StringList{"monday"}, StartTime: "00:00", EndTime: "23:59", Action: action},
	}
}

// aMonday is a fixed instant that falls inside every windowsWithAction test
// window above (Monday, within 00:00-23:59).
var aMonday = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

func TestEngine_EffectivePolicyCombinations(t *testing.T) {
	cases := []struct {
		name            string
		scheduleAction  string
		jellyfinEnabled bool
		jellyfinAction  string
		jellyfinActive  bool
		want            string
	}{
		{"both disabled/none", ActionNone, false, "", false, ActionNone},
		{"schedule none, jellyfin enabled but idle", ActionNone, true, ActionThrottle, false, ActionNone},
		{"schedule throttle, jellyfin disabled", ActionThrottle, false, "", false, ActionThrottle},
		{"schedule stop, jellyfin disabled", ActionStop, false, "", false, ActionStop},
		{"schedule none, jellyfin throttle active", ActionNone, true, ActionThrottle, true, ActionThrottle},
		{"schedule none, jellyfin stop active", ActionNone, true, ActionStop, true, ActionStop},
		{"schedule throttle, jellyfin throttle active", ActionThrottle, true, ActionThrottle, true, ActionThrottle},
		{"schedule throttle, jellyfin stop active", ActionThrottle, true, ActionStop, true, ActionStop},
		{"schedule stop, jellyfin throttle active", ActionStop, true, ActionThrottle, true, ActionStop},
		{"schedule stop, jellyfin enabled but idle", ActionStop, true, ActionThrottle, false, ActionStop},
		{"schedule enabled with no window active, jellyfin enabled but idle", ActionNone, true, ActionStop, false, ActionNone},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cfg := throttleCfg(c.jellyfinEnabled, c.jellyfinAction, 0)
			engine := New(cfg, windowsWithAction(c.scheduleAction), &fakeChecker{active: c.jellyfinActive})
			engine.clock = func() time.Time { return aMonday }
			// Simulate a completed poll without starting a goroutine.
			engine.jellyfinActive.Store(c.jellyfinActive)

			got := engine.Effective()
			if got.Action != c.want {
				t.Errorf("Effective().Action = %q, want %q (schedule=%q jellyfin_enabled=%v jellyfin_action=%q jellyfin_active=%v)",
					got.Action, c.want, c.scheduleAction, c.jellyfinEnabled, c.jellyfinAction, c.jellyfinActive)
			}
		})
	}
}

func TestEngine_JellyfinCheckerNil_AlwaysContributesNone(t *testing.T) {
	cfg := throttleCfg(true, ActionStop, 0) // enabled in config, but no checker available
	engine := New(cfg, windowsWithAction(ActionThrottle), nil)
	engine.clock = func() time.Time { return aMonday }

	got := engine.Effective()
	if got.JellyfinAction != ActionNone {
		t.Errorf("expected JellyfinAction none with nil checker, got %q", got.JellyfinAction)
	}
	if got.Action != ActionThrottle {
		t.Errorf("expected schedule's throttle to still apply, got %q", got.Action)
	}
}

func TestEngine_IsStopped(t *testing.T) {
	cfg := throttleCfg(false, "", 0)
	engine := New(cfg, windowsWithAction(ActionStop), nil)
	engine.clock = func() time.Time { return aMonday }

	if !engine.IsStopped() {
		t.Error("expected IsStopped to be true")
	}
}

func TestEngine_StartPollingWithTicks_UpdatesStateOnEachTick(t *testing.T) {
	cfg := throttleCfg(true, ActionThrottle, 0)
	checker := &fakeChecker{active: false}
	engine := New(cfg, nil, checker)
	engine.clock = func() time.Time { return aMonday }

	ticks := make(chan time.Time)
	engine.startPollingWithTicks(context.Background(), ticks, func() {})
	defer engine.Stop()

	// Synchronous initial poll already happened inside startPollingWithTicks.
	if got := engine.Effective().JellyfinAction; got != ActionNone {
		t.Fatalf("expected none right after start (idle), got %q", got)
	}

	checker.active = true
	ticks <- time.Now()

	deadline := time.After(2 * time.Second)
	for {
		if engine.Effective().JellyfinAction == ActionThrottle {
			break
		}
		select {
		case <-deadline:
			t.Fatal("timed out waiting for poller to observe active playback")
		case <-time.After(time.Millisecond):
		}
	}
}

func TestEngine_StartPolling_NoopWhenJellyfinDisabled(t *testing.T) {
	cfg := throttleCfg(false, "", 0)
	engine := New(cfg, nil, &fakeChecker{active: true})

	engine.StartPolling(context.Background(), time.Millisecond)
	defer engine.Stop()

	// No poller started, so the checker's active state must never surface.
	time.Sleep(10 * time.Millisecond)
	if got := engine.Effective().JellyfinAction; got != ActionNone {
		t.Errorf("expected none with polling disabled, got %q", got)
	}
}

type staticReader struct {
	data []byte
	pos  int
}

func (r *staticReader) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n := copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

func TestWrapReader_StopAbortsWithinOneRead(t *testing.T) {
	cfg := throttleCfg(false, "", 0)
	engine := New(cfg, windowsWithAction(ActionStop), nil)
	engine.clock = func() time.Time { return aMonday }

	underlying := &staticReader{data: make([]byte, 1024)}
	wrapped := engine.WrapReader(context.Background(), underlying)

	buf := make([]byte, 64)
	n, err := wrapped.Read(buf)
	if n != 0 || !errors.Is(err, ErrStoppedByPolicy) {
		t.Fatalf("expected immediate ErrStoppedByPolicy, got n=%d err=%v", n, err)
	}
}

func TestWrapReader_NoneAllowsFullRead(t *testing.T) {
	cfg := throttleCfg(false, "", 0)
	engine := New(cfg, nil, nil)
	engine.clock = func() time.Time { return aMonday }

	underlying := &staticReader{data: []byte("hello world")}
	wrapped := engine.WrapReader(context.Background(), underlying)

	got, err := io.ReadAll(wrapped)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != "hello world" {
		t.Errorf("got %q, want %q", got, "hello world")
	}
}

func TestWrapReader_ThrottleSetsSharedLimiterToConfiguredRate(t *testing.T) {
	cfg := throttleCfg(false, "", 8) // 8 kbps = 1000 bytes/sec
	engine := New(cfg, windowsWithAction(ActionThrottle), nil)
	engine.clock = func() time.Time { return aMonday }

	underlying := &staticReader{data: []byte("hello")}
	wrapped := engine.WrapReader(context.Background(), underlying)

	buf := make([]byte, 5)
	if _, err := wrapped.Read(buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got, want := engine.limiter.Limit(), kbpsToBytesPerSecond(8); got != want {
		t.Errorf("expected shared limiter rate %v bytes/sec while throttled, got %v", want, got)
	}
}

func TestWrapReader_NoneLeavesSharedLimiterUnlimited(t *testing.T) {
	cfg := throttleCfg(false, "", 8) // configured but not currently applicable
	engine := New(cfg, nil, nil)     // no active schedule window -> policy none
	engine.clock = func() time.Time { return aMonday }

	underlying := &staticReader{data: []byte("hello")}
	wrapped := engine.WrapReader(context.Background(), underlying)

	buf := make([]byte, 5)
	if _, err := wrapped.Read(buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := engine.limiter.Limit(); got != rate.Inf {
		t.Errorf("expected unlimited shared limiter under none policy, got %v", got)
	}
}

func TestWrapReader_ThrottleActuallyDelaysBeyondBurst(t *testing.T) {
	// A very low rate (1 byte/sec) makes the delay observable without a
	// slow test: the first read exactly drains the initial burst
	// allowance (no wait), then a second read must wait for tokens to
	// refill at 1/sec, which a short context deadline catches - proving
	// WaitN actually enforces the configured rate rather than being a
	// no-op.
	cfg := throttleCfg(false, "", 0)
	engine := New(cfg, windowsWithAction(ActionThrottle), nil)
	engine.clock = func() time.Time { return aMonday }
	engine.throttleRate = 1 // 1 byte/sec, well below the 64KiB burst

	data := make([]byte, limiterBurst+10)
	underlying := &staticReader{data: data}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	wrapped := engine.WrapReader(ctx, underlying)

	first := make([]byte, limiterBurst)
	n, err := wrapped.Read(first)
	if err != nil || n != limiterBurst {
		t.Fatalf("expected the first read to drain the burst instantly, got n=%d err=%v", n, err)
	}

	second := make([]byte, 10)
	_, err = wrapped.Read(second)
	if err == nil {
		t.Fatal("expected the second read to be blocked by the rate limiter until the context deadline")
	}
}

func TestWrapReader_TransitionsToStopMidTransfer(t *testing.T) {
	cfg := throttleCfg(false, "", 0)
	windows := windowsWithAction(ActionNone)
	engine := New(cfg, windows, nil)

	stopped := false
	engine.clock = func() time.Time {
		if stopped {
			return time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC) // still Monday, irrelevant: schedule empty
		}
		return aMonday
	}
	// Simulate a policy that flips to stop by swapping the schedule windows
	// the Engine holds mid-transfer (as if a later evaluation observed a
	// newly-active stop window).
	underlying := &staticReader{data: make([]byte, 128)}
	wrapped := engine.WrapReader(context.Background(), underlying)

	buf := make([]byte, 32)
	if _, err := wrapped.Read(buf); err != nil {
		t.Fatalf("unexpected error on first read: %v", err)
	}

	engine.windows = windowsWithAction(ActionStop)
	stopped = true

	n, err := wrapped.Read(buf)
	if n != 0 || !errors.Is(err, ErrStoppedByPolicy) {
		t.Fatalf("expected ErrStoppedByPolicy after policy transitions to stop, got n=%d err=%v", n, err)
	}
}
