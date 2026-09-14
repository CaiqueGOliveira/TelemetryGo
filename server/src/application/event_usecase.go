package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/CaiqueGOliveira/TelemetryGo/src/application/dtos"
	"github.com/CaiqueGOliveira/TelemetryGo/src/domain"
	"github.com/CaiqueGOliveira/TelemetryGo/src/domain/messages"
	r "github.com/CaiqueGOliveira/TelemetryGo/src/domain/repository"
	"github.com/google/uuid"
)

const EventsChannel = "telemetry:events"

type EventUsecase struct {
	repo      r.EventRepository
	publisher messages.EventPublisher
}

func NewEventUsecase(repo r.EventRepository, publisher messages.EventPublisher) *EventUsecase {
	return &EventUsecase{
		repo:      repo,
		publisher: publisher,
	}
}

func (uc *EventUsecase) Ingest(requests []*dtos.EventIngestRequestDto, userID string) (int, int, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var accepted int
	var published int
	var firstErr error

	for _, req := range requests {
		wg.Add(1)
		go func(req *dtos.EventIngestRequestDto) {
			defer wg.Done()

			event, err := buildEvent(req, userID)
			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				if firstErr == nil {
					firstErr = err
				}
				return
			}

			if err := uc.repo.Save(event); err != nil {
				if firstErr == nil {
					firstErr = err
				}
				return
			}

			accepted++

			payload, err := json.Marshal(dtos.ToEventResponseDto(event))
			if err != nil {
				slog.Error("failed to marshal event for publish", "event_id", event.Id, "error", err)
				return
			}

			if err := uc.publisher.Publish(context.Background(), EventsChannel, payload); err != nil {
				slog.Error("failed to publish event", "channel", EventsChannel, "error", err)
				return
			}

			published++
		}(req)
	}

	wg.Wait()

	if firstErr != nil {
		return 0, 0, firstErr
	}

	return accepted, published, nil
}

func (uc *EventUsecase) List(userID string, filter r.EventFilter, limit int, offset int) ([]*domain.Event, error) {
	return uc.repo.List(userID, filter, limit, offset)
}

func (uc *EventUsecase) Delete(userID string, id uuid.UUID) error {
	return uc.repo.Delete(userID, id)
}

func (uc *EventUsecase) Subscribe(ctx context.Context) (<-chan []byte, func(), error) {
	return uc.publisher.Subscribe(ctx, EventsChannel)
}

func buildEvent(req *dtos.EventIngestRequestDto, userID string) (*domain.Event, error) {
	id := uuid.New()
	if req.ID != "" {
		parsedID, err := uuid.Parse(req.ID)
		if err != nil {
			return nil, fmt.Errorf("invalid event id: %w", err)
		}
		id = parsedID
	}

	if err := validateRequiredEventFields(req); err != nil {
		return nil, err
	}

	timestamp := time.Now()
	if req.Timestamp != "" {
		parsed, err := time.Parse(time.RFC3339, req.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("invalid event timestamp: %w", err)
		}
		timestamp = parsed
	}

	event, err := domain.NewEvent(
		id,
		req.Type,
		req.Service,
		req.Message,
		req.Severity,
		timestamp,
	)
	if err != nil {
		return nil, err
	}

	event.UserId = userID
	return event, nil
}

func validateRequiredEventFields(req *dtos.EventIngestRequestDto) error {
	switch {
	case req.Type == "":
		return fmt.Errorf("event type is required")
	case req.Service == "":
		return fmt.Errorf("event service is required")
	case req.Message == "":
		return fmt.Errorf("event message is required")
	default:
		return nil
	}
}
