package bootstrap

import (
	"github.com/CaiqueGOliveira/TelemetryGo/src/application"
	"github.com/CaiqueGOliveira/TelemetryGo/src/controllers"
	"github.com/CaiqueGOliveira/TelemetryGo/src/infra/repositories"
	"github.com/CaiqueGOliveira/TelemetryGo/src/infra/routes"
	"github.com/gin-gonic/gin"
)

func AppBootstrap() *gin.Engine {
	repo := repositories.NewUserRepository()

	usecase := application.NewCreateUserUsecase(repo)

	userController := controllers.NewUserController(usecase)

	router := routes.SetupRouter(userController)

	return router
}