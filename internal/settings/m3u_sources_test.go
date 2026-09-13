package settings

import (
	"errors"
	"testing"

	"github.com/glefebvre/stalkeer/internal/config"
)

func configWithSources(sources ...config.M3USourceConfig) *config.Config {
	return &config.Config{M3U: config.M3UConfig{Sources: sources}}
}

func TestOriginSources_EmptyWhenNoneConfigured(t *testing.T) {
	setupTestDB(t)
	config.SetConfig(configWithSources())

	if got := OriginSources(); len(got) != 0 {
		t.Errorf("expected 0 origin sources, got %d", len(got))
	}
}

func TestOriginSources_ReturnsConfiguredEntries(t *testing.T) {
	setupTestDB(t)
	config.SetConfig(configWithSources(config.M3USourceConfig{
		Name:     "a",
		FilePath: "/tmp/a.m3u",
		Download: config.M3UDownloadConfig{URL: "http://a.example.com"},
	}))

	got := OriginSources()
	if len(got) != 1 {
		t.Fatalf("expected 1 origin source, got %d", len(got))
	}
	if got[0].IsRuntime {
		t.Error("origin source must not be reported as runtime")
	}
	if got[0].Source.Name != "a" {
		t.Errorf("expected name 'a', got %q", got[0].Source.Name)
	}
}

func TestEffectiveEntries_OriginOnly(t *testing.T) {
	setupTestDB(t)
	config.SetConfig(configWithSources(config.M3USourceConfig{Name: "a", FilePath: "/tmp/a.m3u"}))

	entries := EffectiveEntries()
	if len(entries) != 1 || entries[0].Source.Name != "a" || entries[0].IsRuntime {
		t.Fatalf("expected 1 origin-only entry 'a', got %+v", entries)
	}
}

func TestEffectiveEntries_OverrideReplacesOrigin(t *testing.T) {
	setupTestDB(t)
	config.SetConfig(configWithSources(config.M3USourceConfig{
		Name:     "b",
		FilePath: "/tmp/b-origin.m3u",
		Download: config.M3UDownloadConfig{URL: "http://origin.example.com"},
	}))

	if _, err := SetSource("b", M3USourceInput{FilePath: "/tmp/b-override.m3u", URL: "http://override.example.com"}); err != nil {
		t.Fatalf("SetSource failed: %v", err)
	}

	entries := EffectiveEntries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if !entries[0].IsRuntime {
		t.Error("expected the override to be reported as runtime")
	}
	if entries[0].Source.FilePath != "/tmp/b-override.m3u" || entries[0].Source.Download.URL != "http://override.example.com" {
		t.Errorf("expected override values, got %+v", entries[0].Source)
	}
}

func TestEffectiveEntries_RuntimeOnlyAddition(t *testing.T) {
	setupTestDB(t)
	config.SetConfig(configWithSources())

	if _, err := SetSource("c", M3USourceInput{FilePath: "/tmp/c.m3u", URL: "http://c.example.com"}); err != nil {
		t.Fatalf("SetSource failed: %v", err)
	}

	entries := EffectiveEntries()
	if len(entries) != 1 || entries[0].Source.Name != "c" || !entries[0].IsRuntime {
		t.Fatalf("expected 1 runtime-only entry 'c', got %+v", entries)
	}
}

func TestEffectiveEntries_EmptyBothSides(t *testing.T) {
	setupTestDB(t)
	config.SetConfig(configWithSources())

	if entries := EffectiveEntries(); len(entries) != 0 {
		t.Errorf("expected empty effective list, got %d entries", len(entries))
	}
}

func TestEffectiveEntries_MixedOriginAndRuntime(t *testing.T) {
	setupTestDB(t)
	config.SetConfig(configWithSources(
		config.M3USourceConfig{Name: "a", FilePath: "/tmp/a.m3u"},
		config.M3USourceConfig{Name: "b", FilePath: "/tmp/b-origin.m3u"},
	))

	if _, err := SetSource("b", M3USourceInput{FilePath: "/tmp/b-override.m3u"}); err != nil {
		t.Fatalf("SetSource failed: %v", err)
	}
	if _, err := SetSource("c", M3USourceInput{FilePath: "/tmp/c.m3u"}); err != nil {
		t.Fatalf("SetSource failed: %v", err)
	}

	byName := make(map[string]M3USourceEntry)
	for _, e := range EffectiveEntries() {
		byName[e.Source.Name] = e
	}

	if len(byName) != 3 {
		t.Fatalf("expected 3 effective sources, got %d", len(byName))
	}
	if byName["a"].IsRuntime {
		t.Error("'a' should remain origin")
	}
	if !byName["b"].IsRuntime || byName["b"].Source.FilePath != "/tmp/b-override.m3u" {
		t.Errorf("'b' should be the runtime override, got %+v", byName["b"])
	}
	if !byName["c"].IsRuntime {
		t.Error("'c' should be runtime-only")
	}
}

func TestDeleteSource_RuntimeOverrideRevertsToOrigin(t *testing.T) {
	setupTestDB(t)
	config.SetConfig(configWithSources(config.M3USourceConfig{Name: "b", FilePath: "/tmp/b-origin.m3u"}))

	if _, err := SetSource("b", M3USourceInput{FilePath: "/tmp/b-override.m3u"}); err != nil {
		t.Fatalf("SetSource failed: %v", err)
	}
	if err := DeleteSource("b"); err != nil {
		t.Fatalf("DeleteSource failed: %v", err)
	}

	entries := EffectiveEntries()
	if len(entries) != 1 || entries[0].IsRuntime || entries[0].Source.FilePath != "/tmp/b-origin.m3u" {
		t.Errorf("expected reverted origin entry, got %+v", entries)
	}
}

func TestDeleteSource_RuntimeOnlyRemovesEntirely(t *testing.T) {
	setupTestDB(t)
	config.SetConfig(configWithSources())

	if _, err := SetSource("c", M3USourceInput{FilePath: "/tmp/c.m3u"}); err != nil {
		t.Fatalf("SetSource failed: %v", err)
	}
	if err := DeleteSource("c"); err != nil {
		t.Fatalf("DeleteSource failed: %v", err)
	}

	if entries := EffectiveEntries(); len(entries) != 0 {
		t.Errorf("expected 'c' fully removed, got %+v", entries)
	}
}

func TestDeleteSource_NoRuntimeSourceIsNotFound(t *testing.T) {
	setupTestDB(t)
	config.SetConfig(configWithSources(config.M3USourceConfig{Name: "a", FilePath: "/tmp/a.m3u"}))

	err := DeleteSource("a")
	if !errors.Is(err, ErrM3USourceNotFound) {
		t.Errorf("expected ErrM3USourceNotFound deleting an origin-only name, got %v", err)
	}
	// The origin source must be unaffected.
	if entries := EffectiveEntries(); len(entries) != 1 || entries[0].Source.Name != "a" {
		t.Errorf("origin source must be unaffected by the rejected delete, got %+v", entries)
	}

	err = DeleteSource("does-not-exist")
	if !errors.Is(err, ErrM3USourceNotFound) {
		t.Errorf("expected ErrM3USourceNotFound deleting an unknown name, got %v", err)
	}
}

func TestSetSource_ReStoringReplacesInPlace(t *testing.T) {
	setupTestDB(t)
	config.SetConfig(configWithSources())

	if _, err := SetSource("c", M3USourceInput{FilePath: "/tmp/c-v1.m3u", RetentionCount: 3}); err != nil {
		t.Fatalf("SetSource failed: %v", err)
	}
	if _, err := SetSource("c", M3USourceInput{FilePath: "/tmp/c-v2.m3u", RetentionCount: 7}); err != nil {
		t.Fatalf("second SetSource failed: %v", err)
	}

	entries := EffectiveEntries()
	if len(entries) != 1 {
		t.Fatalf("expected exactly 1 entry for 'c' (no duplicate), got %d", len(entries))
	}
	if entries[0].Source.FilePath != "/tmp/c-v2.m3u" || entries[0].Source.Download.RetentionCount != 7 {
		t.Errorf("expected the second store's values, got %+v", entries[0].Source)
	}
}

func TestSetSource_PasswordOmittedKeepsCurrent(t *testing.T) {
	setupTestDB(t)
	config.SetConfig(configWithSources())

	pw := "s3cret"
	if _, err := SetSource("c", M3USourceInput{FilePath: "/tmp/c.m3u", AuthPassword: &pw}); err != nil {
		t.Fatalf("SetSource failed: %v", err)
	}

	// Omit AuthPassword entirely (nil) on the next store: current password must survive.
	if _, err := SetSource("c", M3USourceInput{FilePath: "/tmp/c-v2.m3u", RetentionCount: 5}); err != nil {
		t.Fatalf("second SetSource failed: %v", err)
	}

	entries := EffectiveEntries()
	if len(entries) != 1 || !entries[0].HasAuthPassword() {
		t.Fatalf("expected password to survive omission, got %+v", entries)
	}
	if entries[0].Source.Download.AuthPassword != "s3cret" {
		t.Errorf("expected preserved password 's3cret', got %q", entries[0].Source.Download.AuthPassword)
	}
}

func TestSetSource_ExplicitEmptyPasswordClearsIt(t *testing.T) {
	setupTestDB(t)
	config.SetConfig(configWithSources())

	pw := "s3cret"
	if _, err := SetSource("c", M3USourceInput{FilePath: "/tmp/c.m3u", AuthPassword: &pw}); err != nil {
		t.Fatalf("SetSource failed: %v", err)
	}

	empty := ""
	if _, err := SetSource("c", M3USourceInput{FilePath: "/tmp/c.m3u", AuthPassword: &empty}); err != nil {
		t.Fatalf("second SetSource failed: %v", err)
	}

	entries := EffectiveEntries()
	if len(entries) != 1 || entries[0].HasAuthPassword() {
		t.Errorf("expected password cleared, got %+v", entries)
	}
}
