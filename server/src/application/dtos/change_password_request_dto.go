package dtos

type ChangePasswordRequestDto struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}