package repositories

import (
	"sort"
	"sync"

	"github.com/CaiqueGOliveira/TelemetryGo/src/domain"
	"github.com/CaiqueGOliveira/TelemetryGo/src/domain/repository"
	"github.com/google/uuid"
)

type InMemoryEventRepository struct {
	mu     sync.Mutex
	events []*domain.Event
}

func NewInMemoryEventRepository() *InMemoryEventRepository {
	return &InMemoryEventRepository{}
}

func (r *InMemoryEventRepository) Save(event *domain.Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.events = append(r.events, event)
	return nil
}

func (r *InMemoryEventRepository) List(userID string, filter repository.EventFilter, limit int, offset int) ([]*domain.Event, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	result := make([]*domain.Event, 0)
	for _, event := range r.events {
		if event.UserId != userID {
			continue
		}
		if filter.Severity != "" && event.Severity.String() != filter.Severity {
			continue
		}
		if filter.Type != "" && event.Type != filter.Type {
			continue
		}
		if filter.Service != "" && event.Service != filter.Service {
			continue
		}
		if filter.Start != nil && event.Timestamp.Before(*filter.Start) {
			continue
		}
		if filter.End != nil && event.Timestamp.After(*filter.End) {
			continue
		}
		result = append(result, event)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp.After(result[j].Timestamp)
	})

	return paginate(result, limit, offset), nil
}

func (r *InMemoryEventRepository) Delete(userID string, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, event := range r.events {
		if event.UserId == userID && event.Id == id {
			r.events = append(r.events[:i], r.events[i+1:]...)
			return nil
		}
	}

	return repository.ErrNotFound
}

func paginate[T any](items []T, limit int, offset int) []T {
	if offset >= len(items) {
		return []T{}
	}
	start := offset
	end := len(items)
	if limit > 0 && offset+limit < end {
		end = offset + limit
	}
	return items[start:end]
}