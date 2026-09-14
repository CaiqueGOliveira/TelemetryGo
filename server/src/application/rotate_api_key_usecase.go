package application

import (
	r "github.com/CaiqueGOliveira/TelemetryGo/src/domain/repository"
	"github.com/google/uuid"
)

type RotateApiKeyUsecase struct {
	repo r.UserRepository
}

func NewRotateApiKeyUsecase(repo r.UserRepository) *RotateApiKeyUsecase {
	return &RotateApiKeyUsecase{
		repo: repo,
	}
}

func (uc *RotateApiKeyUsecase) Execute(userID string) (string, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return "", err
	}

	user, err := uc.repo.FindById(id)
	if err != nil {
		return "", err
	}

	user.RegenerateApiKey()

	if err := uc.repo.Update(user); err != nil {
		return "", err
	}

	return user.ApiKey, nil
}