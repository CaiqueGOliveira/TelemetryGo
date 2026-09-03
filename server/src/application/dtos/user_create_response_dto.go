package dtos

import "github.com/CaiqueGOliveira/TelemetryGo/src/domain"

type UserCreateResponseDto struct {
	User         *domain.User
	JwtAccess    string
	JwtRefresh   string
}

func NewUserCreateResponseDto(user *domain.User, jwtAccess string, jwtRefresh string) *UserCreateResponseDto {
	return &UserCreateResponseDto{
		User:       user,
		JwtAccess:  jwtAccess,
		JwtRefresh: jwtRefresh,
	}
}
