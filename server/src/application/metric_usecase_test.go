package application

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/CaiqueGOliveira/TelemetryGo/src/application/dtos"
	r "github.com/CaiqueGOliveira/TelemetryGo/src/domain/repository"
	"github.com/CaiqueGOliveira/TelemetryGo/src/infra/repositories"
	"github.com/google/uuid"
)

func newMetricUsecase() (*MetricUsecase, *repositories.InMemoryMetricRepository, *fakePublisher) {
	repo := repositories.NewInMemoryMetricRepository()
	pub := newFakePublisher()
	return NewMetricUsecase(repo, pub), repo, pub
}

func TestMetricIngestSuccess(t *testing.T) {
	uc, repo, pub := newMetricUsecase()

	metricID := uuid.New()
	requests := []*dtos.MetricIngestRequestDto{
		{ID: metricID.String(), Name: "cpu_usage", Service: "api-gateway", Value: "42.5", Unit: "%", Status: "ok", Timestamp: "2026-01-01T10:00:00Z"},
		{Name: "latency", Service: "auth", Value: "120", Unit: "ms", Status: "warn"},
	}

	accepted, published, err := uc.Ingest(requests, "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if accepted != 2 {
		t.Errorf("expected 2 accepted, got %d", accepted)
	}
	if published != 2 {
		t.Errorf("expected 2 published, got %d", published)
	}

	metrics, err := repo.List("user-1", r.MetricFilter{}, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(metrics) != 2 {
		t.Fatalf("expected 2 saved metrics, got %d", len(metrics))
	}

	for _, ch := range pub.publishedChannels() {
		if ch != MetricsChannel {
			t.Errorf("expected publish on %s, got %s", MetricsChannel, ch)
		}
	}

	pub.mu.Lock()
	defer pub.mu.Unlock()
	for _, msg := range pub.published {
		var payload struct {
			Id    string  `json:"id"`
			Value float64 `json:"value"`
		}
		if err := json.Unmarshal(msg.payload, &payload); err != nil {
			t.Fatalf("unexpected payload decode: %v", err)
		}
		if payload.Value == 0 && payload.Id != metricID.String() {
			// one of them is the explicit-id metric; ensure its id shows up
		}
	}
	foundExplicitID := false
	for _, msg := range pub.published {
		var payload struct {
			Id string `json:"id"`
		}
		_ = json.Unmarshal(msg.payload, &payload)
		if payload.Id == metricID.String() {
			foundExplicitID = true
		}
	}
	if !foundExplicitID {
		t.Error("expected published payload to contain explicit metric id")
	}
}

func TestMetricIngestInvalidValue(t *testing.T) {
	uc, _, _ := newMetricUsecase()

	requests := []*dtos.MetricIngestRequestDto{
		{ID: uuid.New().String(), Name: "cpu", Service: "api", Value: "not-a-number", Status: "ok"},
	}

	accepted, published, err := uc.Ingest(requests, "user-1")
	if err == nil {
		t.Fatal("expected error for non-numeric value")
	}
	if accepted != 0 || published != 0 {
		t.Errorf("expected 0/0 on error, got %d/%d", accepted, published)
	}
}

func TestMetricIngestInvalidStatus(t *testing.T) {
	uc, _, _ := newMetricUsecase()

	requests := []*dtos.MetricIngestRequestDto{
		{ID: uuid.New().String(), Name: "cpu", Service: "api", Value: "10", Status: "exploded"},
	}

	if _, _, err := uc.Ingest(requests, "user-1"); err == nil {
		t.Fatal("expected error for invalid status")
	}
}

func TestMetricIngestMissingValue(t *testing.T) {
	uc, _, _ := newMetricUsecase()

	requests := []*dtos.MetricIngestRequestDto{
		{Name: "cpu", Service: "api", Value: "", Status: "ok"},
	}

	if _, _, err := uc.Ingest(requests, "user-1"); err == nil {
		t.Fatal("expected error for missing value")
	}
}

func TestMetricListFilters(t *testing.T) {
	uc, repo, _ := newMetricUsecase()

	base := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	statuses := []string{"ok", "crit", "warn"}
	for i, status := range statuses {
		metric, err := buildMetric(&dtos.MetricIngestRequestDto{
			ID:        uuid.New().String(),
			Name:      "cpu_usage",
			Service:   "api-gateway",
			Value:     "10",
			Unit:      "%",
			Status:    status,
			Timestamp: base.Add(time.Duration(i) * time.Hour).Format(time.RFC3339),
		}, "user-1")
		if err != nil {
			t.Fatal(err)
		}
		if err := repo.Save(metric); err != nil {
			t.Fatal(err)
		}
	}

	crit, err := uc.List("user-1", r.MetricFilter{Status: "crit"}, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(crit) != 1 || crit[0].Status.String() != "crit" {
		t.Errorf("expected 1 crit metric, got %d", len(crit))
	}

	// name filter applies to all saved (all cpu_usage)
	named, err := uc.List("user-1", r.MetricFilter{Name: "cpu_usage"}, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(named) != 3 {
		t.Errorf("expected 3 metrics, got %d", len(named))
	}
}

func TestMetricDelete(t *testing.T) {
	uc, repo, _ := newMetricUsecase()

	id := uuid.New()
	metric, err := buildMetric(&dtos.MetricIngestRequestDto{
		ID: id.String(), Name: "cpu", Service: "api", Value: "10", Unit: "%", Status: "ok",
	}, "user-1")
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.Save(metric); err != nil {
		t.Fatal(err)
	}

	if err := uc.Delete("user-1", id); err != nil {
		t.Fatalf("unexpected delete error: %v", err)
	}
	if err := uc.Delete("user-1", id); err == nil {
		t.Fatal("expected error on second delete")
	}
	if err := uc.Delete("other", id); err == nil {
		t.Fatal("expected error when deleting another tenant's metric")
	}
}

func TestMetricSubscribe(t *testing.T) {
	uc, _, pub := newMetricUsecase()

	ch, cleanup, err := uc.Subscribe(context.Background())
	if err != nil {
		t.Fatalf("unexpected subscribe error: %v", err)
	}
	if ch == nil {
		t.Fatal("expected non-nil channel")
	}

	payload := []byte(`{"id":"x","name":"cpu","value":10}`)
	if err := pub.Publish(context.Background(), MetricsChannel, payload); err != nil {
		t.Fatal(err)
	}

	select {
	case got := <-ch:
		if string(got) != string(payload) {
			t.Errorf("expected payload through channel, got %s", got)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for subscribed payload")
	}

	cleanup()
}
