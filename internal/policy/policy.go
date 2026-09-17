// Package policy combines the weekly bandwidth schedule
// (internal/settings) and Jellyfin active-playback detection
// (internal/external/jellyfin) into one effective download policy, and
// enforces it - as a shared rate limit or a full stop - across the
// download and resume-downloads commands. See the adaptive-download-
// throttling spec and design.md's "A single PolicyEngine".
package policy

import (
	"context"
	"errors"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/time/rate"

	"github.com/glefebvre/stalkeer/internal/config"
	"github.com/glefebvre/stalkeer/internal/models"
	"github.com/glefebvre/stalkeer/internal/settings"
)

// Action values, ordered from least to most restrictive. Re-exported from
// internal/settings so callers need not import both packages.
const (
	ActionNone     = settings.ActionNone
	ActionThrottle = settings.ActionThrottle
	ActionStop     = settings.ActionStop
)

// ErrStoppedByPolicy is returned by a reader obtained from Engine.WrapReader
// when the effective policy is ActionStop. It is a distinct, non-failure
// outcome: see the adaptive-download-throttling spec's "Policy Abort Is Not
// a Failure".
var ErrStoppedByPolicy = errors.New("transfer stopped by download policy")

// defaultPollInterval is used by StartPolling when the configured interval
// is not positive.
const defaultPollInterval = 20 * time.Second

// limiterBurst bounds the largest single WaitN request the shared limiter
// must accept. io.Copy's write-to-file fast path reads in 32KB chunks; this
// is kept comfortably above that so a single chunk is never rejected
// outright, while staying small enough that the initial full-burst
// allowance doesn't meaningfully delay when a real throttle rate kicks in.
const limiterBurst = 64 * 1024 // 64 KiB

// JellyfinChecker resolves whether Jellyfin currently has active playback,
// already fail-open on any error. Satisfied by *jellyfin.Client's
// ActivePlayback method; kept as a narrow interface so the Engine and its
// tests do not depend on a real HTTP client.
type JellyfinChecker interface {
	ActivePlayback(ctx context.Context) bool
}

// Effective is the effective policy at one instant, along with which
// signal(s) are contributing to it. See the adaptive-download-throttling
// spec's "effective policy" read endpoint (task 5.2).
type Effective struct {
	Action         string
	ScheduleAction string
	JellyfinAction string
}

// Engine combines the weekly bandwidth schedule and the Jellyfin
// active-playback signal into one effective policy, and owns the shared
// rate.Limiter enforcing it. One Engine is constructed per
// download/resume-downloads invocation and shared by every worker of that
// run, so "most restrictive wins" is evaluated consistently across all
// concurrent workers at any instant, and Jellyfin is never polled more than
// once per interval regardless of parallelism. See design.md's "Decisions".
type Engine struct {
	windows []models.BandwidthScheduleWindow
	clock   func() time.Time

	jellyfinEnabled bool
	jellyfinAction  string
	checker         JellyfinChecker
	jellyfinActive  atomic.Bool

	limiter      *rate.Limiter
	throttleRate rate.Limit

	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// New constructs an Engine from the effective configuration and the
// currently stored schedule windows. checker may be nil - meaning Jellyfin
// active-playback detection is unavailable (e.g. Jellyfin integration
// disabled entirely) - in which case the Jellyfin signal always contributes
// ActionNone, matching "Configurable Jellyfin Action"'s "Detection
// disabled" scenario. The returned Engine's rate limiter starts unlimited;
// call StartPolling to begin tracking live Jellyfin state.
func New(cfg *config.Config, windows []models.BandwidthScheduleWindow, checker JellyfinChecker) *Engine {
	e := &Engine{
		windows:         windows,
		clock:           time.Now,
		jellyfinEnabled: checker != nil && cfg.Jellyfin.PlaybackCheckEnabled,
		jellyfinAction:  cfg.Jellyfin.PlaybackAction,
		checker:         checker,
		limiter:         rate.NewLimiter(rate.Inf, limiterBurst),
	}
	e.throttleRate = kbpsToBytesPerSecond(cfg.Downloads.ThrottleRateKbps)
	return e
}

func kbpsToBytesPerSecond(kbps int) rate.Limit {
	if kbps <= 0 {
		return rate.Inf
	}
	return rate.Limit(float64(kbps) * 1000 / 8)
}

// StartPolling launches the background Jellyfin poller (see design.md's
// "single background goroutine"): it re-checks Jellyfin's active-playback
// state once immediately, then every interval, caching the fail-open
// result. It is a no-op when Jellyfin detection is disabled, or when
// already started. Call Stop when the run finishes.
func (e *Engine) StartPolling(ctx context.Context, interval time.Duration) {
	if !e.jellyfinEnabled || e.cancel != nil {
		return
	}
	if interval <= 0 {
		interval = defaultPollInterval
	}

	ticker := time.NewTicker(interval)
	e.startPollingWithTicks(ctx, ticker.C, ticker.Stop)
}

// startPollingWithTicks is StartPolling's internal implementation, taking
// an injectable tick source so tests can drive the poll loop deterministically
// without waiting on a real timer.
func (e *Engine) startPollingWithTicks(ctx context.Context, ticks <-chan time.Time, stopSource func()) {
	pollCtx, cancel := context.WithCancel(ctx)
	e.cancel = cancel

	// Poll once synchronously so the first Effective() call after this
	// returns already reflects live Jellyfin state, rather than waiting a
	// full interval. See "Jellyfin Poll Interval".
	e.pollJellyfinOnce(pollCtx)

	e.wg.Add(1)
	go func() {
		defer e.wg.Done()
		defer stopSource()
		for {
			select {
			case <-pollCtx.Done():
				return
			case <-ticks:
				e.pollJellyfinOnce(pollCtx)
			}
		}
	}()
}

func (e *Engine) pollJellyfinOnce(ctx context.Context) {
	e.jellyfinActive.Store(e.checker.ActivePlayback(ctx))
}

// Stop halts the background Jellyfin poller, if running, and waits for it
// to exit before returning.
func (e *Engine) Stop() {
	if e.cancel != nil {
		e.cancel()
		e.wg.Wait()
	}
}

// Effective computes the current effective policy: the weekly schedule
// (evaluated in-memory against the windows loaded at construction, no I/O -
// see "Schedule Evaluation Is Not Delayed By Jellyfin Polling") combined
// with the last-polled Jellyfin signal via most-restrictive-wins. See
// "Effective Policy Combination".
func (e *Engine) Effective() Effective {
	schedule := settings.ActiveScheduleAction(e.windows, e.clock())
	jf := e.jellyfinContribution()
	return Effective{
		Action:         settings.MostRestrictive(schedule, jf),
		ScheduleAction: schedule,
		JellyfinAction: jf,
	}
}

// Action is a convenience for Effective().Action.
func (e *Engine) Action() string {
	return e.Effective().Action
}

// IsStopped reports whether the current effective policy is ActionStop.
func (e *Engine) IsStopped() bool {
	return e.Action() == ActionStop
}

func (e *Engine) jellyfinContribution() string {
	if !e.jellyfinEnabled {
		return ActionNone
	}
	if e.jellyfinActive.Load() {
		return e.jellyfinAction
	}
	return ActionNone
}

// WrapReader wraps r with a policy-aware reader that, before each read,
// aborts with ErrStoppedByPolicy if the effective policy is ActionStop, and
// otherwise applies the shared rate limiter so the configured throttle rate
// is an aggregate cap across every concurrently active transfer rather than
// enforced per-transfer. See "Single Shared Throttle Rate", "No Cap Under a
// None Policy", and "Aborting an In-Progress Transfer on Stop".
func (e *Engine) WrapReader(ctx context.Context, r io.Reader) io.Reader {
	return &policyReader{ctx: ctx, engine: e, underlying: r}
}

type policyReader struct {
	ctx        context.Context
	engine     *Engine
	underlying io.Reader
}

func (pr *policyReader) Read(p []byte) (int, error) {
	action := pr.engine.Action()
	if action == ActionStop {
		return 0, ErrStoppedByPolicy
	}

	n, err := pr.underlying.Read(p)
	if n > 0 {
		limiter := pr.engine.limiterForAction(action)
		if waitErr := limiter.WaitN(pr.ctx, n); waitErr != nil {
			return n, waitErr
		}
	}
	return n, err
}

// limiterForAction returns the shared limiter with its rate synced to
// action: unlimited under none, the configured shared rate under throttle.
// See design.md's "SetLimit() is called whenever the effective policy's
// rate changes".
func (e *Engine) limiterForAction(action string) *rate.Limiter {
	if action == ActionThrottle {
		e.limiter.SetLimit(e.throttleRate)
	} else {
		e.limiter.SetLimit(rate.Inf)
	}
	return e.limiter
}
