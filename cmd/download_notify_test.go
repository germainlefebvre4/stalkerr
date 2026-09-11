package main

import (
	"errors"
	"testing"

	"github.com/glefebvre/stalkeer/internal/notifier"
	"github.com/stretchr/testify/require"
)

func TestNotifyBuildStreamsFailure_BuildsExpectedEvent(t *testing.T) {
	fake := &fakeNotifier{}

	notifyBuildStreamsFailure(fake, errors.New("radarr: connection refused"))

	require.Len(t, fake.events, 1)
	event := fake.events[0]
	require.Equal(t, notifier.Critical, event.Severity)
	require.Contains(t, event.Message, "radarr: connection refused")
}

func TestNotifyDownloadRunResult_NoFailuresSkipsNotification(t *testing.T) {
	fake := &fakeNotifier{}

	notifyDownloadRunResult(fake, &downloadStats{Total: 5, Downloaded: 5, Failed: 0})

	require.Empty(t, fake.events, "a run with zero failures must not send a notification")
}

func TestNotifyDownloadRunResult_WithFailuresSendsEventWithCounts(t *testing.T) {
	fake := &fakeNotifier{}

	notifyDownloadRunResult(fake, &downloadStats{Total: 5, Downloaded: 3, Failed: 2})

	require.Len(t, fake.events, 1)
	event := fake.events[0]
	require.Equal(t, notifier.Warning, event.Severity)
	require.Contains(t, event.Message, "Total: 5")
	require.Contains(t, event.Message, "Downloaded: 3")
	require.Contains(t, event.Message, "Failed: 2")
}
