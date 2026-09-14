package valueobjects

import "testing"

func TestCreateSeverityValid(t *testing.T) {
	for _, want := range []string{"info", "warning", "critical"} {
		severity, err := CreateSeverity(want)
		if err != nil {
			t.Fatalf("unexpected error for %q: %v", want, err)
		}
		if severity.String() != want {
			t.Errorf("expected %q, got %q", want, severity.String())
		}
	}
}

func TestCreateSeverityInvalid(t *testing.T) {
	for _, candidate := range []string{"", "error", "exploded", "INFO"} {
		if _, err := CreateSeverity(candidate); err == nil {
			t.Errorf("expected error for %q", candidate)
		}
	}
}
