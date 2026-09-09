package dtos

type LoginDto struct {
	JwtAccess  string
	JwtRefresh string
	ApiKey     string
}

func NewLoginDto(jwtAccess string, jwtRefresh string, apiKey string) *LoginDto {
	return &LoginDto{
		JwtAccess:  jwtAccess,
		JwtRefresh: jwtRefresh,
		ApiKey:     apiKey,
	}
}
