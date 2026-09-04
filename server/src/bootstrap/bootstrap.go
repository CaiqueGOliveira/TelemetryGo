package bootstrap

import (
	"os"
	"time"

	"github.com/CaiqueGOliveira/TelemetryGo/src/application"
	"github.com/CaiqueGOliveira/TelemetryGo/src/controllers"
	"github.com/CaiqueGOliveira/TelemetryGo/src/infra/auth"
	"github.com/CaiqueGOliveira/TelemetryGo/src/infra/repositories"
	"github.com/CaiqueGOliveira/TelemetryGo/src/infra/routes"
	"github.com/gin-gonic/gin"
)

func AppBootstrap() *gin.Engine {
	repo := repositories.NewUserRepository()

	refreshExpiration := time.Hour * 24 * 7

	jwtProvider := auth.NewJwtService(
		os.Getenv("JWT_SECRET"),
		refreshExpiration,
		time.Hour/6,
	)

	createUserUsecase := application.NewCreateUserUsecase(repo, jwtProvider)
	loginUsecase := application.NewLoginUsecase(jwtProvider, repo)

	userController := controllers.NewUserController(createUserUsecase, loginUsecase, refreshExpiration, jwtProvider)

	router := routes.SetupRouter(userController, jwtProvider)

	return router
}
