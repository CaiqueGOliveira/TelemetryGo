package auth

import (
	"errors"
	"fmt"
	"time"

	t "github.com/CaiqueGOliveira/TelemetryGo/src/application/interfaces"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JwtService struct {
	secret            []byte
	refreshExpiration time.Duration
	accessExpiration  time.Duration
}

func NewJwtService(
	secret string,
	refreshExpiration time.Duration,
	accessExpiration time.Duration,
) *JwtService {
	return &JwtService{
		secret:            []byte(secret),
		refreshExpiration: refreshExpiration,
		accessExpiration:  accessExpiration,
	}
}

func (jw *JwtService) GenerateToken(userID uuid.UUID, tokenType string) (string, error) {
	var token *jwt.Token
	now := time.Now()

	switch tokenType {
	case t.TokenTypeAccess:
		token = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": userID,
			"exp": now.Add(jw.accessExpiration).Unix(),
			"iat": now.Unix(),
			"typ": tokenType,
		})
	case t.TokenTypeRefresh:
		token = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": userID,
			"exp": now.Add(jw.refreshExpiration).Unix(),
			"iat": now.Unix(),
			"typ": tokenType,
		})
	default:
		return "", fmt.Errorf("unknowm token type: %s", tokenType)
	}

	tokenString, err := token.SignedString(jw.secret)
	if err != nil {
		return "", errors.New("error assigning jwt token")
	}

	return tokenString, nil
}

func (jw *JwtService) VerifyToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return jw.secret, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid jwt token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid jwt token")
	}

	return claims, nil
}
