package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewMetricValid(t *testing.T) {
	metric, err := NewMetric(uuid.New(), "cpu_usage", "api-gateway", 42.5, "%", "crit", time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if metric.Name != "cpu_usage" {
		t.Errorf("unexpected name %s", metric.Name)
	}
	if metric.Value != 42.5 {
		t.Errorf("expected value 42.5, got %f", metric.Value)
	}
	if metric.Status.String() != "crit" {
		t.Errorf("expected status crit, got %s", metric.Status.String())
	}
}

func TestNewMetricInvalidStatus(t *testing.T) {
	if _, err := NewMetric(uuid.New(), "cpu", "api", 10, "%", "exploded", time.Now()); err == nil {
		t.Fatal("expected error for invalid status")
	}
}

func TestNewMetricEmptyStatus(t *testing.T) {
	if _, err := NewMetric(uuid.New(), "cpu", "api", 10, "%", "", time.Now()); err == nil {
		t.Fatal("expected error for empty status")
	}
}
