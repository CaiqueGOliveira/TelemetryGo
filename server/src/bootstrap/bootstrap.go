package bootstrap

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/CaiqueGOliveira/TelemetryGo/src/application"
	"github.com/CaiqueGOliveira/TelemetryGo/src/config"
	"github.com/CaiqueGOliveira/TelemetryGo/src/controllers"
	"github.com/CaiqueGOliveira/TelemetryGo/src/domain/messages"
	"github.com/CaiqueGOliveira/TelemetryGo/src/domain/repository"
	"github.com/CaiqueGOliveira/TelemetryGo/src/infra/auth"
	"github.com/CaiqueGOliveira/TelemetryGo/src/infra/database"
	"github.com/CaiqueGOliveira/TelemetryGo/src/infra/messaging"
	"github.com/CaiqueGOliveira/TelemetryGo/src/infra/middleware"
	"github.com/CaiqueGOliveira/TelemetryGo/src/infra/repositories"
	"github.com/CaiqueGOliveira/TelemetryGo/src/infra/routes"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/scylladb/gocqlx/v2"
	"gorm.io/gorm"
)

type App struct {
	Router     *gin.Engine
	ServerPort string
	Cleanup    func()
}

func AppBootstrap() (*App, error) {
	cfg := config.Load()

	var cleanups []func()
	cleanup := func() {
		for i := len(cleanups) - 1; i >= 0; i-- {
			cleanups[i]()
		}
	}

	var userRepo repository.UserRepository = repositories.NewUserRepository()
	var db *gorm.DB

	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		postgresDB, err := database.PostgresConnect(dsn, database.PoolConfig{
			MaxOpenConns:    cfg.DBMaxOpenConns,
			MaxIdleConns:    cfg.DBMaxIdleConns,
			ConnMaxLifetime: cfg.DBConnMaxLifetime,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to connect to postgres: %w", err)
		}
		db = postgresDB
		cleanups = append(cleanups, func() {
			if sqlDB, err := postgresDB.DB(); err == nil {
				_ = sqlDB.Close()
			}
		})
		userRepo = repositories.NewPostgresUserRepository(postgresDB)
	}

	var eventRepo repository.EventRepository = repositories.NewInMemoryEventRepository()
	var metricRepo repository.MetricRepository = repositories.NewInMemoryMetricRepository()
	var cassSession gocqlx.Session

	if hosts := os.Getenv("CASSANDRA_HOSTS"); hosts != "" {
		session, err := database.CassandraConnect(strings.Split(hosts, ","))
		if err != nil {
			return nil, fmt.Errorf("failed to connect to cassandra: %w", err)
		}
		if err := repositories.SetupCassandra(session); err != nil {
			session.Close()
			return nil, fmt.Errorf("failed to setup cassandra schema: %w", err)
		}
		cassSession = session
		cleanups = append(cleanups, func() {
			session.Close()
		})
		eventRepo = repositories.NewCassandraEventRepository(session)
		metricRepo = repositories.NewCassandraMetricRepository(session)
	}

	var eventPublisher messages.EventPublisher = newInMemoryPublisher()
	var redisClient *redis.Client

	if redisURL := os.Getenv("REDIS_URL"); redisURL != "" {
		opts, err := redis.ParseURL(redisURL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse redis url: %w", err)
		}
		client := redis.NewClient(opts)
		if err := client.Ping(context.Background()).Err(); err != nil {
			_ = client.Close()
			return nil, fmt.Errorf("failed to connect to redis: %w", err)
		}
		redisClient = client
		cleanups = append(cleanups, func() {
			_ = client.Close()
		})
		eventPublisher = messaging.NewRedisPublisher(client)
	}

	refreshExpiration := cfg.JWTRefreshExpiration

	jwtProvider, err := auth.NewJwtService(
		os.Getenv("JWT_SECRET"),
		refreshExpiration,
		cfg.JWTAccessExpiration,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize jwt service: %w", err)
	}

	createUserUsecase := application.NewCreateUserUsecase(userRepo, jwtProvider)
	loginUsecase := application.NewLoginUsecase(jwtProvider, userRepo)
	getUserUsecase := application.NewGetUserUsecase(userRepo)
	updateUserUsecase := application.NewUpdateUserUsecase(userRepo)
	changePasswordUsecase := application.NewChangePasswordUsecase(userRepo)
	rotateApiKeyUsecase := application.NewRotateApiKeyUsecase(userRepo)
	deleteUserUsecase := application.NewDeleteUserUsecase(userRepo)
	forgotPasswordUsecase := application.NewForgotPasswordUsecase(userRepo, jwtProvider, cfg.PasswordResetExpiration)
	resetPasswordUsecase := application.NewResetPasswordUsecase(userRepo, jwtProvider)

	userController := controllers.NewUserController(
		createUserUsecase,
		loginUsecase,
		getUserUsecase,
		updateUserUsecase,
		changePasswordUsecase,
		rotateApiKeyUsecase,
		deleteUserUsecase,
		forgotPasswordUsecase,
		resetPasswordUsecase,
		refreshExpiration,
		jwtProvider,
	)

	eventUsecase := application.NewEventUsecase(eventRepo, eventPublisher)
	eventController := controllers.NewEventController(eventUsecase)

	metricUsecase := application.NewMetricUsecase(metricRepo, eventPublisher)
	metricController := controllers.NewMetricController(metricUsecase)

	authRateLimit := middleware.RateLimitConfig{
		RPS:   envFloatOr("RATE_LIMIT_RPS", 10),
		Burst: envIntOr("RATE_LIMIT_BURST", 30),
	}
	ingestRateLimit := middleware.RateLimitConfig{
		RPS:   envFloatOr("INGEST_RATE_LIMIT_RPS", 1000),
		Burst: envIntOr("INGEST_RATE_LIMIT_BURST", 10000),
	}

	healthCheck := newHealthCheck(db, cassSession, redisClient)

	router := routes.SetupRouter(userController, eventController, metricController, jwtProvider, userRepo, authRateLimit, ingestRateLimit, healthCheck)

	return &App{
		Router:     router,
		ServerPort: cfg.ServerPort,
		Cleanup:    cleanup,
	}, nil
}

func newHealthCheck(db *gorm.DB, cass gocqlx.Session, redisClient *redis.Client) func(*gin.Context) {
	return func(c *gin.Context) {
		status := http.StatusOK
		services := gin.H{}

		if db != nil {
			var err error
			sqlDB, dbErr := db.DB()
			if dbErr != nil {
				err = dbErr
			} else {
				ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
				defer cancel()
				err = sqlDB.PingContext(ctx)
			}
			services["postgres"] = healthStatus(err, &status)
		}

		if cass.Session != nil {
			ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
			defer cancel()
			err := cass.Query("SELECT now() FROM system.local", nil).WithContext(ctx).Exec()
			services["cassandra"] = healthStatus(err, &status)
		}

		if redisClient != nil {
			ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
			defer cancel()
			err := redisClient.Ping(ctx).Err()
			services["redis"] = healthStatus(err, &status)
		}

		if status != http.StatusOK {
			c.JSON(status, gin.H{
				"status":   "degraded",
				"services": services,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":   "ok",
			"services": services,
		})
	}
}

func healthStatus(err error, status *int) string {
	if err == nil {
		return "ok"
	}
	*status = http.StatusServiceUnavailable
	return "unavailable: " + err.Error()
}

func envFloatOr(key string, def float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return def
}

func envIntOr(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
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
