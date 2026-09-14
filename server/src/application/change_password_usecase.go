package application

import (
	r "github.com/CaiqueGOliveira/TelemetryGo/src/domain/repository"
	"github.com/google/uuid"
)

type ChangePasswordUsecase struct {
	repo r.UserRepository
}

func NewChangePasswordUsecase(repo r.UserRepository) *ChangePasswordUsecase {
	return &ChangePasswordUsecase{
		repo: repo,
	}
}

func (uc *ChangePasswordUsecase) Execute(userID string, currentPassword string, newPassword string) error {
	id, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	user, err := uc.repo.FindById(id)
	if err != nil {
		return err
	}

	if err := user.ChangePassword(currentPassword, newPassword); err != nil {
		return err
	}

	return uc.repo.Update(user)
}