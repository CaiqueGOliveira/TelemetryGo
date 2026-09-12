package routes

import (
	"net/http"

	t "github.com/CaiqueGOliveira/TelemetryGo/src/application/interfaces"
	"github.com/CaiqueGOliveira/TelemetryGo/src/controllers"
	r "github.com/CaiqueGOliveira/TelemetryGo/src/domain/repository"
	"github.com/CaiqueGOliveira/TelemetryGo/src/infra/middleware"
	"github.com/gin-gonic/gin"
)

func SetupRouter(userController *controllers.UserController, eventController *controllers.EventController, metricController *controllers.MetricController, tokenProvider t.TokenProvider, userRepo r.UserRepository, authRateLimit middleware.RateLimitConfig, ingestRateLimit middleware.RateLimitConfig) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.SecurityHeaders())

	api := r.Group("/api/v1")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "success",
			})
		})

		api.POST("/users", middleware.RateLimitMiddleware(authRateLimit.RPS, authRateLimit.Burst), userController.CreateUser)
		api.POST("/login", middleware.RateLimitMiddleware(authRateLimit.RPS, authRateLimit.Burst), userController.Login)
		api.POST("/auth/refresh", middleware.RateLimitMiddleware(authRateLimit.RPS, authRateLimit.Burst), userController.RefreshToken)
		api.POST("/auth/logout", middleware.RateLimitMiddleware(authRateLimit.RPS, authRateLimit.Burst), userController.Logout)

		ingest := api.Group("")
		ingest.Use(middleware.AuthApiKeyMiddleware(userRepo), middleware.RateLimitMiddleware(ingestRateLimit.RPS, ingestRateLimit.Burst))
		{
			ingest.POST("/events", eventController.Ingest)
			ingest.POST("/metrics", metricController.Ingest)
		}

		protected := api.Group("")
		protected.Use(middleware.Auth(tokenProvider), middleware.RateLimitMiddleware(authRateLimit.RPS, authRateLimit.Burst))
		{
			protected.GET("/me", func(c *gin.Context) {
				claims := c.MustGet("claims")
				c.JSON(http.StatusOK, gin.H{"claims": claims})
			})
			protected.GET("/events", eventController.List)
			protected.GET("/events/stream", eventController.Stream)
			protected.GET("/metrics", metricController.List)
			protected.GET("/metrics/stream", metricController.Stream)
		}
	}

	return r
}
