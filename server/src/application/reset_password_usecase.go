package application

import (
	"errors"

	t "github.com/CaiqueGOliveira/TelemetryGo/src/application/interfaces"
	r "github.com/CaiqueGOliveira/TelemetryGo/src/domain/repository"
	"github.com/google/uuid"
)

type ResetPasswordUsecase struct {
	repo  r.UserRepository
	token t.TokenProvider
}

func NewResetPasswordUsecase(repo r.UserRepository, token t.TokenProvider) *ResetPasswordUsecase {
	return &ResetPasswordUsecase{
		repo:  repo,
		token: token,
	}
}

func (uc *ResetPasswordUsecase) Execute(token string, newPassword string) error {
	claims, err := uc.token.VerifyToken(token)
	if err != nil {
		return errors.New("invalid or expired reset token")
	}

	if typ, _ := claims["typ"].(string); typ != t.TokenTypeReset {
		return errors.New("invalid or expired reset token")
	}

	sub, ok := claims["sub"].(string)
	if !ok {
		return errors.New("invalid or expired reset token")
	}

	userID, err := uuid.Parse(sub)
	if err != nil {
		return errors.New("invalid or expired reset token")
	}

	user, err := uc.repo.FindById(userID)
	if err != nil {
		return errors.New("invalid or expired reset token")
	}

	if err := user.ResetPassword(newPassword); err != nil {
		return err
	}

	return uc.repo.Update(user)
}
