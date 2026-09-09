package routes

import (
	"net/http"

	t "github.com/CaiqueGOliveira/TelemetryGo/src/application/interfaces"
	"github.com/CaiqueGOliveira/TelemetryGo/src/controllers"
	r "github.com/CaiqueGOliveira/TelemetryGo/src/domain/repository"
	"github.com/CaiqueGOliveira/TelemetryGo/src/infra/middleware"
	"github.com/gin-gonic/gin"
)

func SetupRouter(userController *controllers.UserController, eventController *controllers.EventController, metricController *controllers.MetricController, tokenProvider t.TokenProvider, userRepo r.UserRepository) *gin.Engine {
	r := gin.Default()

	api := r.Group("/api/v1")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "success",
			})
		})

		api.POST("/users", userController.CreateUser)
		api.POST("/login", userController.Login)
		api.POST("/auth/refresh", userController.RefreshToken)
		api.POST("/auth/logout", userController.Logout)

		ingest := api.Group("")
		ingest.Use(middleware.AuthApiKeyMiddleware(userRepo))
		{
			ingest.POST("/events", eventController.Ingest)
			ingest.POST("/metrics", metricController.Ingest)
		}

		protected := api.Group("")
		protected.Use(middleware.Auth(tokenProvider))
		{
			protected.GET("/me", func(c *gin.Context) {
				claims := c.MustGet("claims")
				c.JSON(http.StatusOK, gin.H{"claims": claims})
			})
			protected.GET("/events", eventController.List)
			protected.GET("/metrics", metricController.List)
		}
	}

	return r
}
