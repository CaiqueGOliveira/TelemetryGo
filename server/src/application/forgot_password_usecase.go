package application

import (
	"errors"
	"time"

	t "github.com/CaiqueGOliveira/TelemetryGo/src/application/interfaces"
	r "github.com/CaiqueGOliveira/TelemetryGo/src/domain/repository"
)

type ForgotPasswordUsecase struct {
	repo            r.UserRepository
	token           t.TokenProvider
	resetExpiration time.Duration
}

func NewForgotPasswordUsecase(repo r.UserRepository, token t.TokenProvider, resetExpiration time.Duration) *ForgotPasswordUsecase {
	return &ForgotPasswordUsecase{
		repo:            repo,
		token:           token,
		resetExpiration: resetExpiration,
	}
}

// Execute gera um token de reset de senha para o email informado.
// Se o email não existir, retorna token vazio sem erro, para não revelar
// quais emails estão cadastrados (mitigação de enumeração de usuários).
func (uc *ForgotPasswordUsecase) Execute(email string) (string, error) {
	user, err := uc.repo.FindByEmail(email)
	if err != nil {
		if errors.Is(err, r.ErrNotFound) {
			return "", nil
		}
		return "", err
	}

	return uc.token.GenerateResetToken(user.Id, uc.resetExpiration)
}
