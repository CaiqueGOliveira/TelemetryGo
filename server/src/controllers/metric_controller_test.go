package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/CaiqueGOliveira/TelemetryGo/src/application"
	"github.com/CaiqueGOliveira/TelemetryGo/src/domain"
	r "github.com/CaiqueGOliveira/TelemetryGo/src/domain/repository"
	"github.com/CaiqueGOliveira/TelemetryGo/src/infra/messaging"
	"github.com/CaiqueGOliveira/TelemetryGo/src/infra/repositories"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type failingMetricRepo struct{}

func (failingMetricRepo) Save(*domain.Metric) error { return nil }
func (failingMetricRepo) List(string, r.MetricFilter, int, int) ([]*domain.Metric, error) {
	return nil, errors.New("repo boom")
}
func (failingMetricRepo) Delete(string, uuid.UUID) error { return errors.New("repo boom") }

func newMetricTestRouter(uc *application.MetricUsecase) *gin.Engine {
	gin.SetMode(gin.TestMode)

	controller := NewMetricController(uc)

	router := gin.New()
	api := router.Group("/api/v1")
	api.Use(func(c *gin.Context) {
		c.Set("user_id", "user-1")
	})
	api.POST("/metrics", controller.Ingest)
	api.GET("/metrics", controller.List)
	api.DELETE("/metrics/:id", controller.Delete)
	return router
}

func newMetricUsecaseWithRepo(repo r.MetricRepository) *application.MetricUsecase {
	return application.NewMetricUsecase(repo, messaging.NewNoOpPublisher())
}

func validMetricBody() string {
	return `[{"id":"00000000-0000-0000-0000-000000000151","name":"cpu_usage","service":"api-gateway","value":"42.5","unit":"%","status":"ok","timestamp":"2026-01-01T10:00:00Z"}]`
}

func TestMetricControllerIngestSuccess(t *testing.T) {
	router := newMetricTestRouter(newMetricUsecaseWithRepo(repositories.NewInMemoryMetricRepository()))

	resp := do(t, router, http.MethodPost, "/api/v1/metrics", validMetricBody())
	if resp.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", resp.Code, resp.Body.String())
	}

	var body struct {
		Accepted  int `json:"accepted"`
		Published int `json:"published"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("unexpected decode: %v", err)
	}
	if body.Accepted != 1 || body.Published != 1 {
		t.Errorf("expected 1/1, got %d/%d", body.Accepted, body.Published)
	}
}

func TestMetricControllerIngestEmptyArray(t *testing.T) {
	router := newMetricTestRouter(newMetricUsecaseWithRepo(repositories.NewInMemoryMetricRepository()))

	resp := do(t, router, http.MethodPost, "/api/v1/metrics", `[]`)
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.Code)
	}
}

func TestMetricControllerIngestInvalidValue(t *testing.T) {
	router := newMetricTestRouter(newMetricUsecaseWithRepo(repositories.NewInMemoryMetricRepository()))

	resp := do(t, router, http.MethodPost, "/api/v1/metrics",
		`[{"name":"cpu","service":"api","value":"abc","status":"ok"}]`)
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.Code)
	}
}

func TestMetricControllerList(t *testing.T) {
	repo := repositories.NewInMemoryMetricRepository()
	router := newMetricTestRouter(newMetricUsecaseWithRepo(repo))

	if resp := do(t, router, http.MethodPost, "/api/v1/metrics", validMetricBody()); resp.Code != http.StatusCreated {
		t.Fatalf("ingest failed: %s", resp.Body.String())
	}

	resp := do(t, router, http.MethodGet, "/api/v1/metrics", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}

	var metrics []struct {
		Value float64 `json:"value"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &metrics); err != nil {
		t.Fatalf("unexpected decode: %v", err)
	}
	if len(metrics) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(metrics))
	}
	if metrics[0].Value != 42.5 {
		t.Errorf("expected value 42.5, got %f", metrics[0].Value)
	}
}

func TestMetricControllerListInvalidEnd(t *testing.T) {
	router := newMetricTestRouter(newMetricUsecaseWithRepo(repositories.NewInMemoryMetricRepository()))

	resp := do(t, router, http.MethodGet, "/api/v1/metrics?end=not-a-time", "")
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.Code)
	}
}

func TestMetricControllerListRepoError(t *testing.T) {
	router := newMetricTestRouter(newMetricUsecaseWithRepo(failingMetricRepo{}))

	resp := do(t, router, http.MethodGet, "/api/v1/metrics", "")
	if resp.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.Code)
	}
}

func TestMetricControllerDelete(t *testing.T) {
	repo := repositories.NewInMemoryMetricRepository()
	router := newMetricTestRouter(newMetricUsecaseWithRepo(repo))

	metricID := "00000000-0000-0000-0000-000000000151"
	if resp := do(t, router, http.MethodPost, "/api/v1/metrics", validMetricBody()); resp.Code != http.StatusCreated {
		t.Fatalf("ingest failed: %s", resp.Body.String())
	}

	resp := do(t, router, http.MethodDelete, "/api/v1/metrics/"+metricID, "")
	if resp.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.Code)
	}
}

func TestMetricControllerDeleteNotFound(t *testing.T) {
	repo := repositories.NewInMemoryMetricRepository()
	router := newMetricTestRouter(newMetricUsecaseWithRepo(repo))

	resp := do(t, router, http.MethodDelete, "/api/v1/metrics/00000000-0000-0000-0000-000000000199", "")
	if resp.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.Code)
	}
}
