package application

import (
	"fmt"

	"github.com/CaiqueGOliveira/TelemetryGo/src/application/dtos"
	t "github.com/CaiqueGOliveira/TelemetryGo/src/application/interfaces"
	domain "github.com/CaiqueGOliveira/TelemetryGo/src/domain"
	r "github.com/CaiqueGOliveira/TelemetryGo/src/domain/repository"
)

type CreateUserUsecase struct {
	repo  r.UserRepository
	token t.TokenProvider
}

func NewCreateUserUsecase(repo r.UserRepository, token t.TokenProvider) *CreateUserUsecase {
	return &CreateUserUsecase{
		repo:  repo,
		token: token,
	}
}

func (uc *CreateUserUsecase) Execute(req *dtos.UserCreateRequestDto) (*dtos.UserCreateResponseDto, error) {
	user, err := (&domain.User{}).CreateUser(req.Email, req.Name, req.Password)
	if err != nil {
		return nil, err
	}

	if err := uc.repo.Save(user); err != nil {
		return nil, err
	}

	jwtAccess, err := uc.token.GenerateToken(user.Id, "access")
	if err != nil {
		return nil, fmt.Errorf("error generating access jwt token: %v", err)
	}

	jwtRefresh, err := uc.token.GenerateToken(user.Id, "refresh")
	if err != nil {
		return nil, fmt.Errorf("error generating refresh jwt token: %v", err)
	}

	return dtos.NewUserCreateResponseDto(user, jwtAccess, jwtRefresh), nil
}
