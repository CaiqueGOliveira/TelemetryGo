package application

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/CaiqueGOliveira/TelemetryGo/src/application/dtos"
	"github.com/CaiqueGOliveira/TelemetryGo/src/domain/messages"
	r "github.com/CaiqueGOliveira/TelemetryGo/src/domain/repository"
	"github.com/CaiqueGOliveira/TelemetryGo/src/infra/repositories"
	"github.com/google/uuid"
)

type fakePublisher struct {
	mu        sync.Mutex
	published []fakeMessage
	channels  map[string]map[chan []byte]struct{}
}

type fakeMessage struct {
	channel string
	payload []byte
}

func newFakePublisher() *fakePublisher {
	return &fakePublisher{channels: map[string]map[chan []byte]struct{}{}}
}

func (p *fakePublisher) Publish(ctx context.Context, channel string, payload []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.published = append(p.published, fakeMessage{channel: channel, payload: payload})
	for ch := range p.channels[channel] {
		select {
		case ch <- payload:
		default:
		}
	}
	return nil
}

func (p *fakePublisher) Subscribe(ctx context.Context, channel string) (<-chan []byte, func(), error) {
	ch := make(chan []byte, 8)
	p.mu.Lock()
	if p.channels[channel] == nil {
		p.channels[channel] = map[chan []byte]struct{}{}
	}
	p.channels[channel][ch] = struct{}{}
	p.mu.Unlock()

	cleanup := func() {
		p.mu.Lock()
		delete(p.channels[channel], ch)
		close(ch)
		p.mu.Unlock()
	}

	return ch, cleanup, nil
}

func (p *fakePublisher) publishedChannels() []string {
	p.mu.Lock()
	defer p.mu.Unlock()
	result := make([]string, 0, len(p.published))
	for _, msg := range p.published {
		result = append(result, msg.channel)
	}
	return result
}

func newEventUsecase() (*EventUsecase, *repositories.InMemoryEventRepository, *fakePublisher) {
	repo := repositories.NewInMemoryEventRepository()
	pub := newFakePublisher()
	return NewEventUsecase(repo, pub), repo, pub
}

func TestEventIngestSuccess(t *testing.T) {
	uc, repo, pub := newEventUsecase()

	eventID := uuid.New()
	requests := []*dtos.EventIngestRequestDto{
		{ID: eventID.String(), Type: "deploy", Service: "api-gateway", Message: "msg1", Severity: "info", Timestamp: "2026-01-01T10:00:00Z"},
		{Type: "rollback", Service: "auth", Message: "msg2", Severity: "warning"},
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

	events, err := repo.List("user-1", r.EventFilter{}, 10, 0)
	if err != nil {
		t.Fatalf("unexpected list error: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 saved events, got %d", len(events))
	}

	for _, ch := range pub.publishedChannels() {
		if ch != EventsChannel {
			t.Errorf("expected publish on %s, got %s", EventsChannel, ch)
		}
	}

	pub.mu.Lock()
	defer pub.mu.Unlock()
	foundID := false
	for _, msg := range pub.published {
		var payload struct {
			Id string `json:"id"`
		}
		if err := json.Unmarshal(msg.payload, &payload); err != nil {
			t.Fatalf("unexpected payload decode: %v", err)
		}
		if payload.Id == eventID.String() {
			foundID = true
		}
	}
	if !foundID {
		t.Error("expected published payload to contain explicit event id")
	}
}

func TestEventIngestInvalidSeverity(t *testing.T) {
	uc, _, _ := newEventUsecase()

	requests := []*dtos.EventIngestRequestDto{
		{ID: uuid.New().String(), Type: "deploy", Service: "api", Message: "msg", Severity: "not-a-severity"},
	}

	accepted, published, err := uc.Ingest(requests, "user-1")
	if err == nil {
		t.Fatal("expected error for invalid severity")
	}
	if accepted != 0 || published != 0 {
		t.Errorf("expected 0/0 on error, got %d/%d", accepted, published)
	}
}

func TestEventIngestMissingField(t *testing.T) {
	uc, _, _ := newEventUsecase()

	requests := []*dtos.EventIngestRequestDto{
		{Type: "deploy", Service: "api", Message: "", Severity: "info"},
	}

	if _, _, err := uc.Ingest(requests, "user-1"); err == nil {
		t.Fatal("expected error for missing message")
	}
}

func TestEventIngestInvalidTimestamp(t *testing.T) {
	uc, _, _ := newEventUsecase()

	requests := []*dtos.EventIngestRequestDto{
		{Type: "deploy", Service: "api", Message: "msg", Severity: "info", Timestamp: "not-a-time"},
	}

	if _, _, err := uc.Ingest(requests, "user-1"); err == nil {
		t.Fatal("expected error for invalid timestamp")
	}
}

func TestEventListFiltersAndPagination(t *testing.T) {
	uc, repo, _ := newEventUsecase()

	base := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	ids := make([]uuid.UUID, 4)
	for i := 0; i < 4; i++ {
		ids[i] = uuid.New()
		event, err := buildEvent(&dtos.EventIngestRequestDto{
			ID:        ids[i].String(),
			Type:      "deploy",
			Service:   "api-gateway",
			Message:   "msg",
			Severity:  "info",
			Timestamp: base.Add(time.Duration(i) * 30 * time.Minute).Format(time.RFC3339),
		}, "user-1")
		if err != nil {
			t.Fatalf("unexpected build error: %v", err)
		}
		if err := repo.Save(event); err != nil {
			t.Fatal(err)
		}
	}

	byService, err := uc.List("user-1", r.EventFilter{Service: "api-gateway"}, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(byService) != 4 {
		t.Errorf("expected 4 filtered events, got %d", len(byService))
	}

	// sorted desc by timestamp: last saved first
	if byService[0].Id != ids[3] {
		t.Errorf("expected most recent first, got %s", byService[0].Id)
	}

	paged, err := uc.List("user-1", r.EventFilter{}, 2, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(paged) != 2 {
		t.Errorf("expected 2 with limit 2, got %d", len(paged))
	}

	paged2, err := uc.List("user-1", r.EventFilter{}, 2, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(paged2) != 2 || paged2[0].Id != ids[1] {
		t.Errorf("unexpected second page: len=%d first=%s", len(paged2), paged2[0].Id)
	}

	otherUser, err := uc.List("other", r.EventFilter{}, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(otherUser) != 0 {
		t.Errorf("expected no events for other user, got %d", len(otherUser))
	}
}

func TestEventDelete(t *testing.T) {
	uc, repo, _ := newEventUsecase()

	id := uuid.New()
	event, err := buildEvent(&dtos.EventIngestRequestDto{
		ID: id.String(), Type: "deploy", Service: "api", Message: "m", Severity: "info",
	}, "user-1")
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.Save(event); err != nil {
		t.Fatal(err)
	}

	if err := uc.Delete("user-1", id); err != nil {
		t.Fatalf("unexpected delete error: %v", err)
	}

	if err := uc.Delete("user-1", id); err == nil {
		t.Fatal("expected error on second delete")
	}

	// tenant isolation
	if err := uc.Delete("other", id); err == nil {
		t.Fatal("expected error when deleting another tenant's event")
	}
}

func TestEventSubscribe(t *testing.T) {
	uc, _, pub := newEventUsecase()

	ch, cleanup, err := uc.Subscribe(context.Background())
	if err != nil {
		t.Fatalf("unexpected subscribe error: %v", err)
	}
	if ch == nil {
		t.Fatal("expected non-nil channel")
	}

	payload := []byte(`{"id":"x","type":"deploy"}`)
	if err := pub.Publish(context.Background(), EventsChannel, payload); err != nil {
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

func TestEventUsecaseImplementsComposition(t *testing.T) {
	var _ messages.EventPublisher = (*fakePublisher)(nil)
}
