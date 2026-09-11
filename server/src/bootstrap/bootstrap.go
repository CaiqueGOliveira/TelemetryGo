package bootstrap

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/CaiqueGOliveira/TelemetryGo/src/application"
	"github.com/CaiqueGOliveira/TelemetryGo/src/controllers"
	"github.com/CaiqueGOliveira/TelemetryGo/src/domain/messages"
	"github.com/CaiqueGOliveira/TelemetryGo/src/domain/repository"
	"github.com/CaiqueGOliveira/TelemetryGo/src/infra/auth"
	"github.com/CaiqueGOliveira/TelemetryGo/src/infra/database"
	"github.com/CaiqueGOliveira/TelemetryGo/src/infra/messaging"
	"github.com/CaiqueGOliveira/TelemetryGo/src/infra/repositories"
	"github.com/CaiqueGOliveira/TelemetryGo/src/infra/routes"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
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

	var eventPublisher messages.EventPublisher = newInMemoryPublisher()

	if redisURL := os.Getenv("REDIS_URL"); redisURL != "" {
		opts, err := redis.ParseURL(redisURL)
		if err != nil {
			log.Fatalf("failed to parse redis url: %v", err)
		}
		client := redis.NewClient(opts)
		if err := client.Ping(context.Background()).Err(); err != nil {
			log.Fatalf("failed to connect to redis: %v", err)
		}
		eventPublisher = messaging.NewRedisPublisher(client)
	}

	refreshExpiration := time.Hour * 24 * 7

	jwtProvider, err := auth.NewJwtService(
		os.Getenv("JWT_SECRET"),
		refreshExpiration,
		time.Hour/6,
	)
	if err != nil {
		log.Fatalf("failed to initialize jwt service: %v", err)
	}

	createUserUsecase := application.NewCreateUserUsecase(userRepo, jwtProvider)
	loginUsecase := application.NewLoginUsecase(jwtProvider, userRepo)

	userController := controllers.NewUserController(createUserUsecase, loginUsecase, refreshExpiration, jwtProvider)

	eventUsecase := application.NewEventUsecase(eventRepo, eventPublisher)
	eventController := controllers.NewEventController(eventUsecase)

	metricUsecase := application.NewMetricUsecase(metricRepo, eventPublisher)
	metricController := controllers.NewMetricController(metricUsecase)

	router := routes.SetupRouter(userController, eventController, metricController, jwtProvider, userRepo)

	return router
}

type inMemoryPublisher struct{}

func newInMemoryPublisher() *inMemoryPublisher {
	return &inMemoryPublisher{}
}

func (p *inMemoryPublisher) Publish(ctx context.Context, channel string, payload []byte) error {
	return nil
}

func (p *inMemoryPublisher) Subscribe(ctx context.Context, channel string) (<-chan []byte, func(), error) {
	ch := make(chan []byte)
	cleanup := func() { close(ch) }
	return ch, cleanup, nil
}
