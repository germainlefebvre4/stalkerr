package main

import (
	"errors"
	"testing"

	"github.com/glefebvre/stalkeer/internal/notifier"
	"github.com/stretchr/testify/require"
)

func TestNotifyPlaylistFetchFailure_BuildsExpectedEvent(t *testing.T) {
	fake := &fakeNotifier{}

	notifyPlaylistFetchFailure(fake, "provider-a", errors.New("connection timed out"))

	require.Len(t, fake.events, 1)
	event := fake.events[0]
	require.Equal(t, notifier.Critical, event.Severity)
	require.Contains(t, event.Message, "provider-a")
	require.Contains(t, event.Message, "connection timed out")
}
