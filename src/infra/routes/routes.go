package routes

import (
	"net/http"

	"github.com/CaiqueGOliveira/TelemetryGo/src/controllers"
	"github.com/gin-gonic/gin"
)

func SetupRouter(userController *controllers.UserController) *gin.Engine {
	r := gin.Default()

	api := r.Group("/api/v1")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "success",
			})
		})

		api.POST("/users", userController.CreateUser)
	}

	return r
}
