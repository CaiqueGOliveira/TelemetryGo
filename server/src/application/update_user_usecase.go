package application

import (
	"github.com/CaiqueGOliveira/TelemetryGo/src/application/dtos"
	r "github.com/CaiqueGOliveira/TelemetryGo/src/domain/repository"
	"github.com/google/uuid"
)

type UpdateUserUsecase struct {
	repo r.UserRepository
}

func NewUpdateUserUsecase(repo r.UserRepository) *UpdateUserUsecase {
	return &UpdateUserUsecase{
		repo: repo,
	}
}

func (uc *UpdateUserUsecase) Execute(userID string, req *dtos.UpdateUserRequestDto) (*dtos.UserResponseDto, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	user, err := uc.repo.FindById(id)
	if err != nil {
		return nil, err
	}

	if err := user.UpdateProfile(req.Name, req.Email); err != nil {
		return nil, err
	}

	if err := uc.repo.Update(user); err != nil {
		return nil, err
	}

	resp := dtos.ToUserResponseDto(user)
	return &resp, nil
}