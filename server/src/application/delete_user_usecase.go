package application

import (
	r "github.com/CaiqueGOliveira/TelemetryGo/src/domain/repository"
	"github.com/google/uuid"
)

type DeleteUserUsecase struct {
	repo r.UserRepository
}

func NewDeleteUserUsecase(repo r.UserRepository) *DeleteUserUsecase {
	return &DeleteUserUsecase{
		repo: repo,
	}
}

func (uc *DeleteUserUsecase) Execute(userID string) error {
	id, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	return uc.repo.Delete(id)
}