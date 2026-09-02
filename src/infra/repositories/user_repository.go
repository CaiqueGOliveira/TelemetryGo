package repositories

import (
	domain "github.com/CaiqueGOliveira/TelemetryGo/src/domain"
)

type UserRepository struct {
	users []*domain.User
}

func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

func (repo *UserRepository) Save(user *domain.User) error {
	repo.users = append(repo.users, user)
	return nil
}