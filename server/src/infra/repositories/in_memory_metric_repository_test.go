package repositories

import (
	"testing"
	"time"

	domain "github.com/CaiqueGOliveira/TelemetryGo/src/domain"
	r "github.com/CaiqueGOliveira/TelemetryGo/src/domain/repository"
	"github.com/google/uuid"
)

func mustMetric(t *testing.T, id, status string, timestamp time.Time, value float64) *domain.Metric {
	t.Helper()
	metric, err := domain.NewMetric(uuid.MustParse(id), "cpu_usage", "api-gateway", value, "%", status, timestamp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	metric.UserId = "user-1"
	return metric
}

func TestInMemoryMetricRepositoryListFiltersAndOrdering(t *testing.T) {
	repo := NewInMemoryMetricRepository()

	base := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	metrics := []*domain.Metric{
		mustMetric(t, "00000000-0000-0000-0000-000000000111", "ok", base, 10),
		mustMetric(t, "00000000-0000-0000-0000-000000000112", "crit", base.Add(time.Hour), 95),
		mustMetric(t, "00000000-0000-0000-0000-000000000113", "ok", base.Add(2*time.Hour), 20),
	}
	for _, metric := range metrics {
		if err := repo.Save(metric); err != nil {
			t.Fatal(err)
		}
	}

	all, err := repo.List("user-1", r.MetricFilter{}, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 3 {
		t.Fatalf("expected 3 metrics, got %d", len(all))
	}
	if all[0].Id != metrics[2].Id {
		t.Errorf("expected most recent first, got %s", all[0].Id)
	}

	crit, err := repo.List("user-1", r.MetricFilter{Status: "crit"}, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(crit) != 1 || crit[0].Id != metrics[1].Id {
		t.Errorf("expected only crit metric, got %d", len(crit))
	}

	window, err := repo.List("user-1", r.MetricFilter{
		Start: timePtr(base),
		End:   timePtr(base.Add(time.Hour)),
	}, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(window) != 2 {
		t.Errorf("expected 2 metrics in window, got %d", len(window))
	}

	other, err := repo.List("other", r.MetricFilter{}, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(other) != 0 {
		t.Errorf("expected no metrics for other user, got %d", len(other))
	}
}

func TestInMemoryMetricRepositoryDelete(t *testing.T) {
	repo := NewInMemoryMetricRepository()

	target := "00000000-0000-0000-0000-000000000151"
	if err := repo.Save(mustMetric(t, target, "ok", time.Now(), 10)); err != nil {
		t.Fatal(err)
	}
	if err := repo.Save(mustMetric(t, "00000000-0000-0000-0000-000000000152", "ok", time.Now(), 20)); err != nil {
		t.Fatal(err)
	}

	if err := repo.Delete("user-1", uuid.MustParse(target)); err != nil {
		t.Fatalf("unexpected delete error: %v", err)
	}

	remaining, err := repo.List("user-1", r.MetricFilter{}, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(remaining) != 1 {
		t.Fatalf("expected 1 remaining metric, got %d", len(remaining))
	}

	if err := repo.Delete("user-1", uuid.MustParse(target)); err != r.ErrNotFound {
		t.Fatalf("expected ErrNotFound on second delete, got %v", err)
	}

	if err := repo.Delete("other", uuid.MustParse("00000000-0000-0000-0000-000000000152")); err != r.ErrNotFound {
		t.Fatalf("expected ErrNotFound deleting another tenant's metric, got %v", err)
	}
}
