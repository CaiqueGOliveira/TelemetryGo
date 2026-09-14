package repositories

import (
	"os"
	"strings"
	"testing"
	"time"

	domain "github.com/CaiqueGOliveira/TelemetryGo/src/domain"
	"github.com/CaiqueGOliveira/TelemetryGo/src/domain/repository"
	"github.com/CaiqueGOliveira/TelemetryGo/src/infra/database"
	"github.com/google/uuid"
	"github.com/scylladb/gocqlx/v2"
)

func cassandraSession(t *testing.T) gocqlx.Session {
	t.Helper()

	hosts := os.Getenv("CASSANDRA_HOSTS")
	if hosts == "" {
		t.Skip("CASSANDRA_HOSTS not set; skipping cassandra integration tests")
	}

	session, err := database.CassandraConnect(strings.Split(hosts, ","))
	if err != nil {
		t.Fatalf("failed to connect to cassandra: %v", err)
	}
	t.Cleanup(session.Close)

	if err := SetupCassandra(session); err != nil {
		t.Fatalf("failed to setup schema: %v", err)
	}

	return session
}

func eventByID(t *testing.T, events []*domain.Event, id uuid.UUID) bool {
	t.Helper()
	for _, event := range events {
		if event.Id == id {
			return true
		}
	}
	return false
}

func TestCassandraEventRepositoryCRUD(t *testing.T) {
	session := cassandraSession(t)
	repo := NewCassandraEventRepository(session)

	userID := uuid.NewString()
	now := time.Now().UTC().Truncate(time.Millisecond)
	event, err := domain.NewEvent(uuid.New(), "deploy", "api-gateway", "field deploy", "info", now.Add(-time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	event.UserId = userID

	if err := repo.Save(event); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	all, err := repo.List(userID, repository.EventFilter{}, 10, 0)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if !eventByID(t, all, event.Id) {
		t.Fatal("saved event not found in list")
	}

	window, err := repo.List(userID, repository.EventFilter{
		Start: timePtr(now.Add(-2 * time.Minute)),
		End:   timePtr(now),
	}, 10, 0)
	if err != nil {
		t.Fatalf("filtered list failed: %v", err)
	}
	if !eventByID(t, window, event.Id) {
		t.Fatal("event not found in time window")
	}

	severity, err := repo.List(userID, repository.EventFilter{Severity: "info"}, 10, 0)
	if err != nil {
		t.Fatalf("severity filter failed: %v", err)
	}
	if !eventByID(t, severity, event.Id) {
		t.Fatal("event not found in severity filter")
	}

	if err := repo.Delete(userID, event.Id); err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	after, err := repo.List(userID, repository.EventFilter{}, 100, 0)
	if err != nil {
		t.Fatalf("list after delete failed: %v", err)
	}
	if eventByID(t, after, event.Id) {
		t.Fatal("event still present after delete")
	}
}

func metricByID(t *testing.T, metrics []*domain.Metric, id uuid.UUID) bool {
	t.Helper()
	for _, metric := range metrics {
		if metric.Id == id {
			return true
		}
	}
	return false
}

func TestCassandraMetricRepositoryCRUD(t *testing.T) {
	session := cassandraSession(t)
	repo := NewCassandraMetricRepository(session)

	userID := uuid.NewString()
	now := time.Now().UTC().Truncate(time.Millisecond)
	metric, err := domain.NewMetric(uuid.New(), "cpu_usage", "api-gateway", 42.5, "%", "ok", now.Add(-time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	metric.UserId = userID

	if err := repo.Save(metric); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	all, err := repo.List(userID, repository.MetricFilter{}, 10, 0)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if !metricByID(t, all, metric.Id) {
		t.Fatal("saved metric not found in list")
	}

	crit, err := repo.List(userID, repository.MetricFilter{Status: "warn"}, 10, 0)
	if err != nil {
		t.Fatalf("status filter failed: %v", err)
	}
	if metricByID(t, crit, metric.Id) {
		t.Fatal("metric should not match warn filter")
	}

	if err := repo.Delete(userID, metric.Id); err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	after, err := repo.List(userID, repository.MetricFilter{}, 100, 0)
	if err != nil {
		t.Fatalf("list after delete failed: %v", err)
	}
	if metricByID(t, after, metric.Id) {
		t.Fatal("metric still present after delete")
	}
}
