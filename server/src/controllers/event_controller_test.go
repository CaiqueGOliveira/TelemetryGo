package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/CaiqueGOliveira/TelemetryGo/src/application"
	"github.com/CaiqueGOliveira/TelemetryGo/src/domain"
	r "github.com/CaiqueGOliveira/TelemetryGo/src/domain/repository"
	"github.com/CaiqueGOliveira/TelemetryGo/src/infra/messaging"
	"github.com/CaiqueGOliveira/TelemetryGo/src/infra/repositories"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type failingEventRepo struct{}

func (failingEventRepo) Save(*domain.Event) error { return nil }
func (failingEventRepo) List(string, r.EventFilter, int, int) ([]*domain.Event, error) {
	return nil, errors.New("repo boom")
}
func (failingEventRepo) Delete(string, uuid.UUID) error { return errors.New("repo boom") }

func newEventTestRouter(uc *application.EventUsecase) *gin.Engine {
	gin.SetMode(gin.TestMode)

	controller := NewEventController(uc)

	router := gin.New()
	api := router.Group("/api/v1")
	api.Use(func(c *gin.Context) {
		c.Set("user_id", "user-1")
	})
	api.POST("/events", controller.Ingest)
	api.GET("/events", controller.List)
	api.DELETE("/events/:id", controller.Delete)
	return router
}

func newEventUsecaseWithRepo(repo r.EventRepository) *application.EventUsecase {
	return application.NewEventUsecase(repo, messaging.NewNoOpPublisher())
}

func do(t *testing.T, router *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func validEventBody() string {
	return `[{"id":"00000000-0000-0000-0000-000000000041","type":"deploy","service":"api-gateway","message":"m","severity":"info","timestamp":"2026-01-01T10:00:00Z"}]`
}

func TestEventControllerIngestSuccess(t *testing.T) {
	router := newEventTestRouter(newEventUsecaseWithRepo(repositories.NewInMemoryEventRepository()))

	resp := do(t, router, http.MethodPost, "/api/v1/events", validEventBody())
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

func TestEventControllerIngestEmptyArray(t *testing.T) {
	router := newEventTestRouter(newEventUsecaseWithRepo(repositories.NewInMemoryEventRepository()))

	resp := do(t, router, http.MethodPost, "/api/v1/events", `[]`)
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.Code)
	}
}

func TestEventControllerIngestMalformedBody(t *testing.T) {
	router := newEventTestRouter(newEventUsecaseWithRepo(repositories.NewInMemoryEventRepository()))

	resp := do(t, router, http.MethodPost, "/api/v1/events", `not json`)
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.Code)
	}
}

func TestEventControllerIngestInvalidSeverity(t *testing.T) {
	router := newEventTestRouter(newEventUsecaseWithRepo(repositories.NewInMemoryEventRepository()))

	resp := do(t, router, http.MethodPost, "/api/v1/events",
		`[{"type":"deploy","service":"api","message":"m","severity":"exploded"}]`)
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.Code)
	}
}

func TestEventControllerList(t *testing.T) {
	repo := repositories.NewInMemoryEventRepository()
	router := newEventTestRouter(newEventUsecaseWithRepo(repo))

	if resp := do(t, router, http.MethodPost, "/api/v1/events", validEventBody()); resp.Code != http.StatusCreated {
		t.Fatalf("ingest failed: %s", resp.Body.String())
	}

	resp := do(t, router, http.MethodGet, "/api/v1/events", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}

	var events []struct {
		Id string `json:"id"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &events); err != nil {
		t.Fatalf("unexpected decode: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
}

func TestEventControllerListInvalidStart(t *testing.T) {
	router := newEventTestRouter(newEventUsecaseWithRepo(repositories.NewInMemoryEventRepository()))

	resp := do(t, router, http.MethodGet, "/api/v1/events?start=not-a-time", "")
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.Code)
	}
}

func TestEventControllerListRepoError(t *testing.T) {
	router := newEventTestRouter(newEventUsecaseWithRepo(failingEventRepo{}))

	resp := do(t, router, http.MethodGet, "/api/v1/events", "")
	if resp.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.Code)
	}
}

func TestEventControllerDelete(t *testing.T) {
	repo := repositories.NewInMemoryEventRepository()
	router := newEventTestRouter(newEventUsecaseWithRepo(repo))

	eventID := "00000000-0000-0000-0000-000000000041"
	if resp := do(t, router, http.MethodPost, "/api/v1/events", validEventBody()); resp.Code != http.StatusCreated {
		t.Fatalf("ingest failed: %s", resp.Body.String())
	}

	resp := do(t, router, http.MethodDelete, "/api/v1/events/"+eventID, "")
	if resp.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", resp.Code, resp.Body.String())
	}
}

func TestEventControllerDeleteInvalidID(t *testing.T) {
	router := newEventTestRouter(newEventUsecaseWithRepo(repositories.NewInMemoryEventRepository()))

	resp := do(t, router, http.MethodDelete, "/api/v1/events/not-a-uuid", "")
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.Code)
	}
}

func TestEventControllerDeleteNotFound(t *testing.T) {
	repo := repositories.NewInMemoryEventRepository()
	router := newEventTestRouter(newEventUsecaseWithRepo(repo))

	resp := do(t, router, http.MethodDelete, "/api/v1/events/00000000-0000-0000-0000-000000000099", "")
	if resp.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.Code)
	}
}

func TestEventControllerDeleteRepoError(t *testing.T) {
	router := newEventTestRouter(newEventUsecaseWithRepo(failingEventRepo{}))

	resp := do(t, router, http.MethodDelete, "/api/v1/events/00000000-0000-0000-0000-000000000041", "")
	if resp.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.Code)
	}
}
