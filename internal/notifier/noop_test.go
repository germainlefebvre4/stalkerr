package notifier

import (
	"context"
	"testing"
)

func TestNoopNotifier_NeverErrors(t *testing.T) {
	n := noopNotifier{}

	err := n.Notify(context.Background(), Event{
		Severity: Critical,
		Title:    "test",
		Message:  "test message",
	})

	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}
