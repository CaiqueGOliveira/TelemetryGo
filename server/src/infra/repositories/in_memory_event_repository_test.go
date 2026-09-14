package repositories

import (
	"testing"
	"time"

	domain "github.com/CaiqueGOliveira/TelemetryGo/src/domain"
	r "github.com/CaiqueGOliveira/TelemetryGo/src/domain/repository"
	"github.com/google/uuid"
)

func timePtr(t time.Time) *time.Time {
	return &t
}

func mustEvent(t *testing.T, id string, severity string, timestamp time.Time) *domain.Event {
	t.Helper()
	event, err := domain.NewEvent(uuid.MustParse(id), "deploy", "api-gateway", "msg", severity, timestamp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	event.UserId = "user-1"
	return event
}

func TestInMemoryEventRepositoryListFiltersAndOrdering(t *testing.T) {
	repo := NewInMemoryEventRepository()

	base := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	events := []*domain.Event{
		mustEvent(t, "00000000-0000-0000-0000-000000000011", "info", base),
		mustEvent(t, "00000000-0000-0000-0000-000000000012", "critical", base.Add(time.Hour)),
		mustEvent(t, "00000000-0000-0000-0000-000000000013", "info", base.Add(2*time.Hour)),
	}
	for _, event := range events {
		if err := repo.Save(event); err != nil {
			t.Fatal(err)
		}
	}

	// ordered desc by timestamp
	all, err := repo.List("user-1", r.EventFilter{}, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 3 {
		t.Fatalf("expected 3 events, got %d", len(all))
	}
	if all[0].Id != events[2].Id {
		t.Errorf("expected most recent first, got %s", all[0].Id)
	}

	// severity filter
	critical, err := repo.List("user-1", r.EventFilter{Severity: "critical"}, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(critical) != 1 || critical[0].Id != events[1].Id {
		t.Errorf("expected only critical event, got %d", len(critical))
	}

	// time range filter
	window, err := repo.List("user-1", r.EventFilter{
		Start: timePtr(base),
		End:   timePtr(base.Add(time.Hour)),
	}, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(window) != 2 {
		t.Errorf("expected 2 events in window, got %d", len(window))
	}

	// tenant isolation
	other, err := repo.List("other", r.EventFilter{}, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(other) != 0 {
		t.Errorf("expected no events for other user, got %d", len(other))
	}
}

func TestInMemoryEventRepositoryPagination(t *testing.T) {
	repo := NewInMemoryEventRepository()

	base := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	for i := 0; i < 5; i++ {
		id := uuid.New()
		event, err := domain.NewEvent(id, "deploy", "api", "msg", "info", base.Add(time.Duration(i)*time.Hour))
		if err != nil {
			t.Fatal(err)
		}
		event.UserId = "user-1"
		if err := repo.Save(event); err != nil {
			t.Fatal(err)
		}
	}

	page1, err := repo.List("user-1", r.EventFilter{}, 2, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(page1) != 2 {
		t.Fatalf("expected 2 on first page, got %d", len(page1))
	}

	page2, err := repo.List("user-1", r.EventFilter{}, 2, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(page2) != 2 {
		t.Fatalf("expected 2 on second page, got %d", len(page2))
	}

	last, err := repo.List("user-1", r.EventFilter{}, 2, 4)
	if err != nil {
		t.Fatal(err)
	}
	if len(last) != 1 {
		t.Fatalf("expected 1 on last page, got %d", len(last))
	}

	outOfRange, err := repo.List("user-1", r.EventFilter{}, 2, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(outOfRange) != 0 {
		t.Fatalf("expected empty page, got %d", len(outOfRange))
	}
}

func TestInMemoryEventRepositoryDelete(t *testing.T) {
	repo := NewInMemoryEventRepository()

	target := "00000000-0000-0000-0000-000000000041"
	if err := repo.Save(mustEvent(t, target, "info", time.Now())); err != nil {
		t.Fatal(err)
	}
	if err := repo.Save(mustEvent(t, "00000000-0000-0000-0000-000000000042", "info", time.Now())); err != nil {
		t.Fatal(err)
	}

	if err := repo.Delete("user-1", uuid.MustParse(target)); err != nil {
		t.Fatalf("unexpected delete error: %v", err)
	}

	remaining, err := repo.List("user-1", r.EventFilter{}, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(remaining) != 1 {
		t.Fatalf("expected 1 remaining event, got %d", len(remaining))
	}

	if err := repo.Delete("user-1", uuid.MustParse(target)); err != r.ErrNotFound {
		t.Fatalf("expected ErrNotFound on second delete, got %v", err)
	}

	// tenant isolation
	if err := repo.Delete("other", uuid.MustParse("00000000-0000-0000-0000-000000000042")); err != r.ErrNotFound {
		t.Fatalf("expected ErrNotFound deleting another tenant's event, got %v", err)
	}
}

func TestPaginateHelper(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}

	if got := paginate(items, 2, 0); len(got) != 2 || got[0] != 1 {
		t.Errorf("unexpected first page: %v", got)
	}
	if got := paginate(items, 2, 4); len(got) != 1 || got[0] != 5 {
		t.Errorf("unexpected last page: %v", got)
	}
	if got := paginate(items, 2, 6); len(got) != 0 {
		t.Errorf("expected empty page, got %v", got)
	}
	if got := paginate(items, 0, 0); len(got) != 5 {
		t.Errorf("expected full slice with limit 0, got %v", got)
	}
}
