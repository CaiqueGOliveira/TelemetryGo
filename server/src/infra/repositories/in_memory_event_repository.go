package repositories

import (
	"sync"

	"github.com/CaiqueGOliveira/TelemetryGo/src/domain"
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

func (r *InMemoryEventRepository) FindAll(userID string) []*domain.Event {
	r.mu.Lock()
	defer r.mu.Unlock()

	result := make([]*domain.Event, 0, len(r.events))
	for _, event := range r.events {
		if event.UserId == userID {
			result = append(result, event)
		}
	}
	return result
}
