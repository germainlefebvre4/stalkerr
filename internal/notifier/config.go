package notifier

import "github.com/glefebvre/stalkeer/internal/config"

// FromConfig builds the Notifier described by cfg. It returns the ntfy
// implementation when notifications and the ntfy channel are enabled with a
// server URL and topic configured, and a noopNotifier otherwise (disabled,
// or enabled but missing required fields) so call sites never need a nil
// check or an "if notifications enabled" branch.
func FromConfig(cfg config.NotificationsConfig) Notifier {
	if !cfg.Enabled {
		return noopNotifier{}
	}

	if !cfg.Ntfy.Enabled || cfg.Ntfy.ServerURL == "" || cfg.Ntfy.Topic == "" {
		return noopNotifier{}
	}

	return newNtfyNotifier(cfg.Ntfy.ServerURL, cfg.Ntfy.Topic, cfg.Ntfy.AuthToken)
}
