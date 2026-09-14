package application

import (
	"github.com/CaiqueGOliveira/TelemetryGo/src/application/dtos"
	r "github.com/CaiqueGOliveira/TelemetryGo/src/domain/repository"
	"github.com/google/uuid"
)

type GetUserUsecase struct {
	repo r.UserRepository
}

func NewGetUserUsecase(repo r.UserRepository) *GetUserUsecase {
	return &GetUserUsecase{
		repo: repo,
	}
}

func (gu *GetUserUsecase) Execute(userID string) (*dtos.UserResponseDto, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	user, err := gu.repo.FindById(id)
	if err != nil {
		return nil, err
	}

	resp := dtos.ToUserResponseDto(user)
	return &resp, nil
}
