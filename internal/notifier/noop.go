package notifier

import "context"

// noopNotifier discards every Event. It is used when notifications are
// disabled, or enabled but missing required channel configuration, so call
// sites never need a nil check or an "if notifications enabled" branch.
type noopNotifier struct{}

// Notify does nothing and always returns nil.
func (noopNotifier) Notify(ctx context.Context, event Event) error {
	return nil
}
