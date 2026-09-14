package interfaces

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
	TokenTypeReset   = "reset-password"
)

type TokenProvider interface {
	GenerateToken(userID uuid.UUID, tokenType string) (string, error)
	GenerateResetToken(userID uuid.UUID, expiration time.Duration) (string, error)
	VerifyToken(tokenString string) (jwt.MapClaims, error)
}
