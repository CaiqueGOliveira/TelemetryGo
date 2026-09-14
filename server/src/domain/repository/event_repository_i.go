package repository

import (
	"github.com/CaiqueGOliveira/TelemetryGo/src/domain"
	"github.com/google/uuid"
)

type EventRepository interface {
	Save(event *domain.Event) error
	List(userID string, filter EventFilter, limit int, offset int) ([]*domain.Event, error)
	Delete(userID string, id uuid.UUID) error
}
