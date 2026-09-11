package notifier

import (
	"testing"

	"github.com/glefebvre/stalkeer/internal/config"
)

func TestFromConfig_Disabled(t *testing.T) {
	n := FromConfig(config.NotificationsConfig{Enabled: false})

	if _, ok := n.(noopNotifier); !ok {
		t.Fatalf("expected noopNotifier when notifications disabled, got %T", n)
	}
}

func TestFromConfig_EnabledMissingNtfyFields(t *testing.T) {
	n := FromConfig(config.NotificationsConfig{
		Enabled: true,
		Ntfy: config.NtfyConfig{
			Enabled: true,
			// ServerURL and Topic left empty
		},
	})

	if _, ok := n.(noopNotifier); !ok {
		t.Fatalf("expected noopNotifier when ntfy is missing required fields, got %T", n)
	}
}

func TestFromConfig_EnabledWithValidNtfy(t *testing.T) {
	n := FromConfig(config.NotificationsConfig{
		Enabled: true,
		Ntfy: config.NtfyConfig{
			Enabled:   true,
			ServerURL: "https://ntfy.sh",
			Topic:     "stalkeer-alerts",
			AuthToken: "secret",
		},
	})

	nt, ok := n.(*ntfyNotifier)
	if !ok {
		t.Fatalf("expected *ntfyNotifier when ntfy is fully configured, got %T", n)
	}
	if nt.serverURL != "https://ntfy.sh" {
		t.Errorf("expected serverURL 'https://ntfy.sh', got %q", nt.serverURL)
	}
	if nt.topic != "stalkeer-alerts" {
		t.Errorf("expected topic 'stalkeer-alerts', got %q", nt.topic)
	}
	if nt.authToken != "secret" {
		t.Errorf("expected authToken 'secret', got %q", nt.authToken)
	}
}
