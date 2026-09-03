package dtos

type LoginDto struct {
	JwtAccess  string
	JwtRefresh string
}

func NewLoginDto(jwtAccess string, jwtRefresh string) *LoginDto {
	return &LoginDto{
		JwtAccess:  jwtAccess,
		JwtRefresh: jwtRefresh,
	}
}
