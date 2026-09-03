package dtos

type UserCreateRequestDto struct {
	Name     string
	Email    string
	Password string
}

func NewUserCreateRequestDto(name string, email string, password string) *UserCreateRequestDto {
	return &UserCreateRequestDto{
		Name:     name,
		Email:    email,
		Password: password,
	}
}
