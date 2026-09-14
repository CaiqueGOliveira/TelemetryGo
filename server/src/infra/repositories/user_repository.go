package repositories

import (
	"errors"
	"sync"

	domain "github.com/CaiqueGOliveira/TelemetryGo/src/domain"
	repository "github.com/CaiqueGOliveira/TelemetryGo/src/domain/repository"
	"github.com/google/uuid"
)

type UserRepository struct {
	mu    sync.Mutex
	users []*domain.User
}

func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

func (repo *UserRepository) Save(user *domain.User) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	if user == nil {
		return errors.New("user is nil")
	}

	for _, u := range repo.users {
		if u.Email.Text() == user.Email.Text() {
			return repository.ErrDuplicateEmail
		}
	}

	repo.users = append(repo.users, user)
	return nil
}

func (repo *UserRepository) Update(user *domain.User) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	found := false
	for _, other := range repo.users {
		if other.Id == user.Id {
			found = true
			continue
		}
		if other.Email.Text() == user.Email.Text() {
			return repository.ErrDuplicateEmail
		}
	}

	if !found {
		return repository.ErrNotFound
	}

	for i := range repo.users {
		if repo.users[i].Id == user.Id {
			repo.users[i] = user
			break
		}
	}

	return nil
}

func (repo *UserRepository) Delete(id uuid.UUID) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	for i, user := range repo.users {
		if user.Id == id {
			repo.users = append(repo.users[:i], repo.users[i+1:]...)
			return nil
		}
	}

	return repository.ErrNotFound
}

func (repo *UserRepository) FindByEmail(email string) (*domain.User, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	for _, user := range repo.users {
		if user.Email.Text() == email {
			return user, nil
		}
	}

	return nil, repository.ErrNotFound
}

func (repo *UserRepository) FindByApiKey(apiKey string) (*domain.User, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	for _, user := range repo.users {
		if user.ApiKey == apiKey {
			return user, nil
		}
	}

	return nil, repository.ErrNotFound
}

func (repo *UserRepository) FindById(id uuid.UUID) (*domain.User, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	for _, user := range repo.users {
		if user.Id == id {
			return user, nil
		}
	}

	return nil, repository.ErrNotFound
}