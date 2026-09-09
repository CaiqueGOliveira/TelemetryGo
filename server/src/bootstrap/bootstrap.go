package bootstrap

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/CaiqueGOliveira/TelemetryGo/src/application"
	"github.com/CaiqueGOliveira/TelemetryGo/src/controllers"
	"github.com/CaiqueGOliveira/TelemetryGo/src/domain/repository"
	"github.com/CaiqueGOliveira/TelemetryGo/src/infra/auth"
	"github.com/CaiqueGOliveira/TelemetryGo/src/infra/database"
	"github.com/CaiqueGOliveira/TelemetryGo/src/infra/repositories"
	"github.com/CaiqueGOliveira/TelemetryGo/src/infra/routes"
	"github.com/gin-gonic/gin"
)

func AppBootstrap() *gin.Engine {
	var userRepo repository.UserRepository = repositories.NewUserRepository()

	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		db, err := database.PostgresConnect(dsn)
		if err != nil {
			log.Fatalf("failed to connect to postgres: %v", err)
		}
		userRepo = repositories.NewPostgresUserRepository(db)
	}

	var eventRepo repository.EventRepository = repositories.NewInMemoryEventRepository()
	var metricRepo repository.MetricRepository = repositories.NewInMemoryMetricRepository()

	if hosts := os.Getenv("CASSANDRA_HOSTS"); hosts != "" {
		session, err := database.CassandraConnect(strings.Split(hosts, ","))
		if err != nil {
			log.Fatalf("failed to connect to cassandra: %v", err)
		}
		if err := repositories.SetupCassandra(session); err != nil {
			log.Fatalf("failed to setup cassandra schema: %v", err)
		}
		eventRepo = repositories.NewCassandraEventRepository(session)
		metricRepo = repositories.NewCassandraMetricRepository(session)
	}

	refreshExpiration := time.Hour * 24 * 7

	jwtProvider := auth.NewJwtService(
		os.Getenv("JWT_SECRET"),
		refreshExpiration,
		time.Hour/6,
	)

	createUserUsecase := application.NewCreateUserUsecase(userRepo, jwtProvider)
	loginUsecase := application.NewLoginUsecase(jwtProvider, userRepo)

	userController := controllers.NewUserController(createUserUsecase, loginUsecase, refreshExpiration, jwtProvider)

	eventUsecase := application.NewEventUsecase(eventRepo)
	eventController := controllers.NewEventController(eventUsecase)

	metricUsecase := application.NewMetricUsecase(metricRepo)
	metricController := controllers.NewMetricController(metricUsecase)

	router := routes.SetupRouter(userController, eventController, metricController, jwtProvider, userRepo)

	return router
}
