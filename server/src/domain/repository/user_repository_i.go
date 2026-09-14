package repository

import (
	"github.com/CaiqueGOliveira/TelemetryGo/src/domain"
	"github.com/google/uuid"
)

type UserRepository interface {
	Save(user *domain.User) error
	Update(user *domain.User) error
	Delete(id uuid.UUID) error
	FindByEmail(email string) (*domain.User, error)
	FindByApiKey(apiKey string) (*domain.User, error)
	FindById(id uuid.UUID) (*domain.User, error)
}
