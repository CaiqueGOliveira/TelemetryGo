package application

import (
	"fmt"

	"github.com/CaiqueGOliveira/TelemetryGo/src/application/dtos"
	t "github.com/CaiqueGOliveira/TelemetryGo/src/application/interfaces"
	r "github.com/CaiqueGOliveira/TelemetryGo/src/domain/repository"
)

type LoginUsecase struct {
	repo  r.UserRepository
	token t.TokenProvider
}

func NewLoginUsecase(token t.TokenProvider, repo r.UserRepository) *LoginUsecase {
	return &LoginUsecase{
		token: token,
		repo:  repo,
	}
}

func (l *LoginUsecase) Execute(email, password string) (*dtos.LoginDto, error) {
	user, err := l.repo.FindByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("user's email not found")
	}

	match, err := user.Hash.Verify(password)
	if err != nil {
		return nil, fmt.Errorf("error verifying password: %v", err)
	}
	if !match {
		return nil, fmt.Errorf("invalid credentials")
	}

	jwtAccess, err := l.token.GenerateToken(user.Id, "access")
	if err != nil {
		return nil, fmt.Errorf("error generating access jwt token: %v", err)
	}

	jwtRefresh, err := l.token.GenerateToken(user.Id, "refresh")
	if err != nil {
		return nil, fmt.Errorf("error generating refresh jwt token: %v", err)
	}

	response := dtos.NewLoginDto(
		jwtAccess,
		jwtRefresh,
	)

	return response, nil
}
