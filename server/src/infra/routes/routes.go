package routes

import (
	"net/http"

	t "github.com/CaiqueGOliveira/TelemetryGo/src/application/interfaces"
	"github.com/CaiqueGOliveira/TelemetryGo/src/controllers"
	r "github.com/CaiqueGOliveira/TelemetryGo/src/domain/repository"
	"github.com/CaiqueGOliveira/TelemetryGo/src/infra/middleware"
	"github.com/gin-gonic/gin"
)

func SetupRouter(userController *controllers.UserController, eventController *controllers.EventController, metricController *controllers.MetricController, tokenProvider t.TokenProvider, userRepo r.UserRepository, authRateLimit middleware.RateLimitConfig, ingestRateLimit middleware.RateLimitConfig, healthCheck func(*gin.Context)) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.SecurityHeaders())

	if healthCheck == nil {
		healthCheck = func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "success",
			})
		}
	}

	api := r.Group("/api/v1")
	{
		api.GET("/health", healthCheck)

		api.POST("/users", middleware.RateLimitMiddleware(authRateLimit.RPS, authRateLimit.Burst), userController.CreateUser)
		api.POST("/login", middleware.RateLimitMiddleware(authRateLimit.RPS, authRateLimit.Burst), userController.Login)
		api.POST("/auth/refresh", middleware.RateLimitMiddleware(authRateLimit.RPS, authRateLimit.Burst), userController.RefreshToken)
		api.POST("/auth/logout", middleware.RateLimitMiddleware(authRateLimit.RPS, authRateLimit.Burst), userController.Logout)
		api.POST("/auth/forgot-password", middleware.RateLimitMiddleware(authRateLimit.RPS, authRateLimit.Burst), userController.ForgotPassword)
		api.POST("/auth/reset-password", middleware.RateLimitMiddleware(authRateLimit.RPS, authRateLimit.Burst), userController.ResetPassword)

		ingest := api.Group("")
		ingest.Use(middleware.AuthApiKeyMiddleware(userRepo), middleware.RateLimitMiddleware(ingestRateLimit.RPS, ingestRateLimit.Burst))
		{
			ingest.POST("/events", eventController.Ingest)
			ingest.POST("/metrics", metricController.Ingest)
		}

		protected := api.Group("")
		protected.Use(middleware.Auth(tokenProvider), middleware.RateLimitMiddleware(authRateLimit.RPS, authRateLimit.Burst))
		{
			protected.GET("/me", userController.Me)
			protected.PUT("/me", userController.UpdateProfile)
			protected.DELETE("/users/me", userController.DeleteAccount)
			protected.POST("/auth/change-password", userController.ChangePassword)
			protected.POST("/auth/rotate-api-key", userController.RotateApiKey)
			protected.GET("/events", eventController.List)
			protected.GET("/events/stream", eventController.Stream)
			protected.DELETE("/events/:id", eventController.Delete)
			protected.GET("/metrics", metricController.List)
			protected.GET("/metrics/stream", metricController.Stream)
			protected.DELETE("/metrics/:id", metricController.Delete)
		}
	}

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "route not found"})
	})
	r.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
	})

	return r
}
