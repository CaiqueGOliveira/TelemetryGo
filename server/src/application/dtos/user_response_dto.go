package dtos

import "github.com/CaiqueGOliveira/TelemetryGo/src/domain"

type UserResponseDto struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	ApiKey string `json:"api_key"`
}

func ToUserResponseDto(user *domain.User) UserResponseDto {
	return UserResponseDto{
		Id:     user.Id.String(),
		Name:   user.Name,
		Email:  user.Email.Text(),
		ApiKey: user.ApiKey,
	}
}
