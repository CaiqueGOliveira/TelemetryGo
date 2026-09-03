package repositories

import (
	"fmt"

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

func (repo *UserRepository) FindByEmail(email string) (*domain.User, error) {
	for _, user := range repo.users {
		if user.Email.Text() == email {
			return user, nil
		}
	}

	return nil, fmt.Errorf("user doesn't exist")
}
