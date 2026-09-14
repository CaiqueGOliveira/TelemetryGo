package bootstrap

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/scylladb/gocqlx/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func newHealthRouter(handler func(*gin.Context)) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/health", handler)
	return router
}

func doHealth(router *gin.Engine) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestHealthCheckOk(t *testing.T) {
	router := newHealthRouter(newHealthCheck(nil, gocqlx.Session{}, nil))

	resp := doHealth(router)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.Code, resp.Body.String())
	}

	var body map[string]interface{}
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["status"] != "ok" {
		t.Errorf("expected status ok, got %v", body["status"])
	}
}

func TestHealthCheckPostgresDegraded(t *testing.T) {
	db := newClosedGormDB(t)
	router := newHealthRouter(newHealthCheck(db, gocqlx.Session{}, nil))

	resp := doHealth(router)
	if resp.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d: %s", resp.Code, resp.Body.String())
	}

	var body map[string]interface{}
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["status"] != "degraded" {
		t.Errorf("expected status degraded, got %v", body["status"])
	}
	services := body["services"].(map[string]interface{})
	if !strings.HasPrefix(services["postgres"].(string), "unavailable") {
		t.Errorf("expected postgres unavailable, got %v", services["postgres"])
	}
}

func TestHealthCheckRedisDegraded(t *testing.T) {
	client := redis.NewClient(&redis.Options{Addr: "127.0.0.1:1"})
	t.Cleanup(func() { _ = client.Close() })

	router := newHealthRouter(newHealthCheck(nil, gocqlx.Session{}, client))

	resp := doHealth(router)
	if resp.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d: %s", resp.Code, resp.Body.String())
	}

	var body map[string]interface{}
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	services := body["services"].(map[string]interface{})
	if !strings.HasPrefix(services["redis"].(string), "unavailable") {
		t.Errorf("expected redis unavailable, got %v", services["redis"])
	}
}

func newClosedGormDB(t *testing.T) *gorm.DB {
	t.Helper()

	sqlDB, err := sql.Open("pgx", "postgres://user:pass@127.0.0.1:1/db")
	if err != nil {
		t.Fatalf("failed to open sql db: %v", err)
	}

	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{DisableAutomaticPing: true})
	if err != nil {
		t.Fatalf("failed to open gorm: %v", err)
	}

	raw, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get underlying db: %v", err)
	}
	if err := raw.Close(); err != nil {
		t.Fatalf("failed to close underlying db: %v", err)
	}

	return db
}

func TestHealthCheckShutdown(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	client := redis.NewClient(&redis.Options{Addr: "127.0.0.1:1"})
	t.Cleanup(func() { _ = client.Close() })

	req := httptest.NewRequest(http.MethodGet, "/health", nil).WithContext(ctx)
	router := newHealthRouter(newHealthCheck(nil, gocqlx.Session{}, client))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}
}
