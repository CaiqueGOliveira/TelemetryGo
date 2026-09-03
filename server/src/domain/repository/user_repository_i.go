package repository

import (
	"github.com/CaiqueGOliveira/TelemetryGo/src/domain"
	vo "github.com/CaiqueGOliveira/TelemetryGo/src/domain"
)

type UserRepository interface {
	Save(user *vo.User) error
	FindByEmail(email string) (*domain.User, error)
}
