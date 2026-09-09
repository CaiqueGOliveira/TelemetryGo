package application

import (
	"sync"
	"time"

	"github.com/CaiqueGOliveira/TelemetryGo/src/application/dtos"
	"github.com/CaiqueGOliveira/TelemetryGo/src/domain"
	r "github.com/CaiqueGOliveira/TelemetryGo/src/domain/repository"
	"github.com/google/uuid"
)

type EventUsecase struct {
	repo r.EventRepository
}

func NewEventUsecase(repo r.EventRepository) *EventUsecase {
	return &EventUsecase{
		repo: repo,
	}
}

func (uc *EventUsecase) Ingest(requests []*dtos.EventIngestRequestDto, userID string) (int, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var accepted int
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
		}(req)
	}

	wg.Wait()

	if firstErr != nil {
		return 0, firstErr
	}

	return accepted, nil
}

func (uc *EventUsecase) List(userID string) []*domain.Event {
	return uc.repo.FindAll(userID)
}

func buildEvent(req *dtos.EventIngestRequestDto, userID string) (*domain.Event, error) {
	id := req.ID
	if id == "" {
		id = uuid.NewString()
	}

	eventType := req.Type
	if eventType == "" {
		eventType = "event"
	}

	timestamp := time.Now()
	if req.Timestamp != "" {
		parsed, err := time.Parse(time.RFC3339, req.Timestamp)
		if err == nil {
			timestamp = parsed
		}
	}

	event, err := domain.NewEvent(
		id,
		eventType,
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
