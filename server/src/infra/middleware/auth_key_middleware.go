package middleware

import (
	"net/http"
	"strings"

	r "github.com/CaiqueGOliveira/TelemetryGo/src/domain/repository"
	"github.com/gin-gonic/gin"
)

func AuthApiKeyMiddleware(repo r.UserRepository) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		apiKey := ctx.GetHeader("X-API-Key")
		if apiKey == "" {
			header := ctx.GetHeader("Authorization")
			if header != "" {
				parts := strings.SplitN(header, " ", 2)
				if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
					apiKey = parts[1]
				}
			}
		}

		if apiKey == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing api key"})
			return
		}

		user, err := repo.FindByApiKey(apiKey)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid api key"})
			return
		}

		ctx.Set("user", user)
		ctx.Set("user_id", user.Id.String())
		ctx.Next()
	}
}
