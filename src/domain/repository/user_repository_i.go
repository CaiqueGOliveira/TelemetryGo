package repository

import (
	vo "github.com/CaiqueGOliveira/TelemetryGo/src/domain"
)

type UserRepository interface {
	Save(user *vo.User) error
}
