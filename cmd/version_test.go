package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVersionString_ReportsInjectedBuildMetadata(t *testing.T) {
	originalVersion, originalCommit, originalDate := version, commit, date
	defer func() { version, commit, date = originalVersion, originalCommit, originalDate }()

	version = "1.2.3"
	commit = "abc1234"
	date = "2026-09-10_12:00:00"

	got := versionString()

	require.Contains(t, got, "1.2.3")
	require.Contains(t, got, "abc1234")
	require.Contains(t, got, "2026-09-10_12:00:00")
}

func TestVersionString_DefaultsWhenUnset(t *testing.T) {
	originalVersion, originalCommit, originalDate := version, commit, date
	defer func() { version, commit, date = originalVersion, originalCommit, originalDate }()

	version, commit, date = "dev", "unknown", "unknown"

	got := versionString()

	require.Contains(t, got, "dev")
	require.Contains(t, got, "unknown")
}
