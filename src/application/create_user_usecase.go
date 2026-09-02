package application

import (
	dto "github.com/CaiqueGOliveira/TelemetryGo/src/application/dtos"
	domain "github.com/CaiqueGOliveira/TelemetryGo/src/domain"
	r "github.com/CaiqueGOliveira/TelemetryGo/src/domain/repository"
)

type CreateUserUsecase struct {
	repo r.UserRepository
}

func NewCreateUserUsecase(repo r.UserRepository) *CreateUserUsecase {
	return &CreateUserUsecase{repo: repo}
}

func (uc *CreateUserUsecase) Execute(dto *dto.UserCreateRequestDto) (*domain.User, error) {
	user, err := (&domain.User{}).CreateUser(dto.Email, dto.Name, dto.Password)
	if err != nil {
		return nil, err
	}

	if err := uc.repo.Save(user); err != nil {
		return nil, err
	}

	return user, nil
}