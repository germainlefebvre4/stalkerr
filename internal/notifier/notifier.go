// Package notifier provides a channel-agnostic push notification abstraction
// used to alert operators of failed or broken scheduled runs (download,
// process, m3u-download) without requiring anyone to check the dashboard or
// pod logs proactively.
package notifier

import "context"

// Severity classifies how urgently an Event should be surfaced to an
// operator.
type Severity string

const (
	// Warning marks a run-with-failures summary: the run completed but some
	// items/errors were recorded.
	Warning Severity = "warning"
	// Critical marks a structural failure: the run could not proceed past a
	// required upstream dependency (e.g. Radarr/Sonarr unreachable, playlist
	// fetch failure, M3U parse failure).
	Critical Severity = "critical"
)

// Event describes a single notification to deliver. Each Notifier
// implementation is responsible for its own formatting (e.g. ntfy
// priority/tags); call sites never branch on channel type.
type Event struct {
	Severity Severity
	Title    string
	Message  string
}

// Notifier delivers Events to an operator through some channel.
type Notifier interface {
	Notify(ctx context.Context, event Event) error
}
