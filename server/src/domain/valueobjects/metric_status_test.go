package valueobjects

import "testing"

func TestCreateMetricStatusValid(t *testing.T) {
	for _, want := range []string{"ok", "warn", "crit"} {
		status, err := CreateMetricStatus(want)
		if err != nil {
			t.Fatalf("unexpected error for %q: %v", want, err)
		}
		if status.String() != want {
			t.Errorf("expected %q, got %q", want, status.String())
		}
	}
}

func TestCreateMetricStatusInvalid(t *testing.T) {
	for _, candidate := range []string{"", "OK", "exploded", "critical"} {
		if _, err := CreateMetricStatus(candidate); err == nil {
			t.Errorf("expected error for %q", candidate)
		}
	}
}
