package dtos

type UpdateUserRequestDto struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}