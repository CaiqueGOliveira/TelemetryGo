package routes

import (
	"net/http"

	t "github.com/CaiqueGOliveira/TelemetryGo/src/application/interfaces"
	"github.com/CaiqueGOliveira/TelemetryGo/src/controllers"
	"github.com/CaiqueGOliveira/TelemetryGo/src/infra/middleware"
	"github.com/gin-gonic/gin"
)

func SetupRouter(userController *controllers.UserController, tokenProvider t.TokenProvider) *gin.Engine {
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
	}

	protected := api.Group("")
	protected.Use(middleware.Auth(tokenProvider))
	{
		protected.GET("/me", func(c *gin.Context) {
			claims := c.MustGet("claims").(interface{})
			c.JSON(http.StatusOK, gin.H{"claims": claims})
		})
	}

	return r
}
