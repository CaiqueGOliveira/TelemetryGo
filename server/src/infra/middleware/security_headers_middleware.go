package middleware

import "github.com/gin-gonic/gin"

func SecurityHeaders() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Header("X-Content-Type-Options", "nosniff")
		ctx.Header("X-Frame-Options", "DENY")
		ctx.Header("Referrer-Policy", "no-referrer")
		ctx.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=(), interest-cohort=()")
		ctx.Header("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		ctx.Next()
	}
}
