package main

import (
	"context"

	"github.com/glefebvre/stalkeer/internal/notifier"
)

// fakeNotifier captures every Event passed to Notify instead of delivering
// it anywhere, so tests can assert on exactly what a helper would have sent.
type fakeNotifier struct {
	events []notifier.Event
}

func (f *fakeNotifier) Notify(ctx context.Context, event notifier.Event) error {
	f.events = append(f.events, event)
	return nil
}
