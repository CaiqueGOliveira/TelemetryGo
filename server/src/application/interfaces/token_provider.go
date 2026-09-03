package interfaces

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenProvider interface {
	GenerateToken(userID uuid.UUID, tokenType string) (string, error)
	VerifyToken(tokenString string) (jwt.MapClaims, error)
}
