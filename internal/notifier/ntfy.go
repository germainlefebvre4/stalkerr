package notifier

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/glefebvre/stalkeer/internal/retry"
)

// ntfyRetryConfig bounds delivery retries to a small, short-lived budget:
// Notify is always the last thing a command does, after its real work is
// already complete or already failed, so it must not add much latency at the
// very end of a cron job.
func ntfyRetryConfig() retry.Config {
	return retry.Config{
		MaxAttempts:       3,
		InitialBackoff:    500 * time.Millisecond,
		MaxBackoff:        5 * time.Second,
		BackoffMultiplier: 2.0,
		JitterFraction:    0.1,
	}
}

// ntfyIsRetryable treats any delivery failure (network error or non-2xx
// response) as worth retrying within the bounded attempt budget.
func ntfyIsRetryable(err error) bool {
	return err != nil
}

// ntfyNotifier posts Events to a configured ntfy (https://ntfy.sh) topic via
// a plain HTTP POST, matching how internal/external clients talk to their
// APIs (hand-rolled net/http, no SDK dependency).
type ntfyNotifier struct {
	serverURL string
	topic     string
	authToken string
	http      *http.Client
	retry     retry.Config
}

// newNtfyNotifier creates a Notifier that publishes to serverURL/topic.
func newNtfyNotifier(serverURL, topic, authToken string) *ntfyNotifier {
	return &ntfyNotifier{
		serverURL: strings.TrimRight(serverURL, "/"),
		topic:     topic,
		authToken: authToken,
		http:      &http.Client{Timeout: 5 * time.Second},
		retry:     ntfyRetryConfig(),
	}
}

// Notify posts event to the configured ntfy topic, retrying a transient
// failure a bounded number of times before giving up.
func (n *ntfyNotifier) Notify(ctx context.Context, event Event) error {
	return retry.Do(ctx, n.retry, func() error {
		return n.publish(ctx, event)
	}, ntfyIsRetryable)
}

func (n *ntfyNotifier) publish(ctx context.Context, event Event) error {
	url := n.serverURL + "/" + n.topic

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(event.Message))
	if err != nil {
		return fmt.Errorf("failed to create ntfy request: %w", err)
	}

	priority, tag := ntfyPriorityAndTag(event.Severity)
	req.Header.Set("Title", event.Title)
	req.Header.Set("Priority", priority)
	req.Header.Set("Tags", tag)
	if n.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+n.authToken)
	}

	resp, err := n.http.Do(req)
	if err != nil {
		return fmt.Errorf("failed to reach ntfy server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ntfy server returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// ntfyPriorityAndTag maps a Severity to the ntfy priority/tag pair used to
// format the notification. Warning (run-with-failures summaries) gets
// default priority with a warning tag; Critical (structural failures) gets
// high priority with a rotating_light tag.
func ntfyPriorityAndTag(s Severity) (priority, tag string) {
	if s == Critical {
		return "high", "rotating_light"
	}
	return "default", "warning"
}
