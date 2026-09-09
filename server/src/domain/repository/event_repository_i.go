package repository

import "github.com/CaiqueGOliveira/TelemetryGo/src/domain"

type EventRepository interface {
	Save(event *domain.Event) error
	FindAll(userID string) []*domain.Event
}