package tmdb

import (
	"errors"
	"testing"

	"github.com/glefebvre/stalkeer/internal/circuitbreaker"
)

// 4.1: CircuitBreakerStatus reflects the client's circuit breaker state after
// it transitions, forced here by simulated failures tripping it open.
func TestClient_CircuitBreakerStatus_ReflectsStateAfterTransition(t *testing.T) {
	client := NewClient(Config{APIKey: "test-key"})

	status := client.CircuitBreakerStatus()
	if status.State != circuitbreaker.StateClosed {
		t.Fatalf("expected initial state closed, got %v", status.State)
	}
	if status.Failures != 0 {
		t.Fatalf("expected initial failures 0, got %d", status.Failures)
	}

	// Default breaker config (see NewClient) trips open after 5 failures.
	for i := 0; i < 5; i++ {
		_ = client.circuitBrk.Execute(func() error {
			return errors.New("simulated failure")
		})
	}

	status = client.CircuitBreakerStatus()
	if status.State != circuitbreaker.StateOpen {
		t.Errorf("expected state open after 5 failures, got %v", status.State)
	}
	if status.Failures != 5 {
		t.Errorf("expected failures 5, got %d", status.Failures)
	}
}
