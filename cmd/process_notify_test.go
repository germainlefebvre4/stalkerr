package main

import (
	"errors"
	"testing"

	"github.com/glefebvre/stalkeer/internal/notifier"
	"github.com/glefebvre/stalkeer/internal/processor"
	"github.com/stretchr/testify/require"
)

func TestNotifyProcessParseFailure_BuildsExpectedEvent(t *testing.T) {
	fake := &fakeNotifier{}

	notifyProcessParseFailure(fake, errors.New("malformed m3u header"))

	require.Len(t, fake.events, 1)
	event := fake.events[0]
	require.Equal(t, notifier.Critical, event.Severity)
	require.Contains(t, event.Message, "malformed m3u header")
}

func TestNotifyProcessRunResult_NoErrorsSkipsNotification(t *testing.T) {
	fake := &fakeNotifier{}

	notifyProcessRunResult(fake, &processor.Statistics{Processed: 10, Errors: 0})

	require.Empty(t, fake.events, "a run with zero errors must not send a notification")
}

func TestNotifyProcessRunResult_WithErrorsSendsEventWithCounts(t *testing.T) {
	fake := &fakeNotifier{}

	notifyProcessRunResult(fake, &processor.Statistics{Processed: 10, Errors: 3})

	require.Len(t, fake.events, 1)
	event := fake.events[0]
	require.Equal(t, notifier.Warning, event.Severity)
	require.Contains(t, event.Message, "Processed: 10")
	require.Contains(t, event.Message, "Errors: 3")
}
