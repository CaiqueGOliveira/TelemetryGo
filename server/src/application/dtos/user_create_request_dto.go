package dtos

type UserCreateRequestDto struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func NewUserCreateRequestDto(name string, email string, password string) *UserCreateRequestDto {
	return &UserCreateRequestDto{
		Name:     name,
		Email:    email,
		Password: password,
	}
}
