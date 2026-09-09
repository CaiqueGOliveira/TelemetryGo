package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	domain "github.com/CaiqueGOliveira/TelemetryGo/src/domain"
	vo "github.com/CaiqueGOliveira/TelemetryGo/src/domain/valueobjects"
	"github.com/CaiqueGOliveira/TelemetryGo/src/infra/repositories"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func createTestUser(apiKey string) (*domain.User, error) {
	email, err := vo.CreateEmail("test@example.com")
	if err != nil {
		return nil, err
	}

	passwordHash := &vo.PasswordHashed{}
	hashed, err := passwordHash.CreateHash("Password123!")
	if err != nil {
		return nil, err
	}

	return &domain.User{
		Id:     uuid.New(),
		Email:  email,
		Name:   "Test User",
		Hash:   hashed,
		ApiKey: apiKey,
	}, nil
}

func setupApiKey(apiKey string) *gin.Engine {
	gin.SetMode(gin.TestMode)

	repo := repositories.NewUserRepository()
	user, err := createTestUser(apiKey)
	if err != nil {
		panic(err)
	}
	_ = repo.Save(user)

	r := gin.New()
	protected := r.Group("/")
	protected.Use(AuthApiKeyMiddleware(repo))
	protected.POST("/ingest", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"ok": true})
	})

	return r
}

func TestApiKeyValid(t *testing.T) {
	r := setupApiKey("sk-test-123")

	req := httptest.NewRequest(http.MethodPost, "/ingest", nil)
	req.Header.Set("X-API-Key", "sk-test-123")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestApiKeyViaBearerHeader(t *testing.T) {
	r := setupApiKey("sk-bearer-456")

	req := httptest.NewRequest(http.MethodPost, "/ingest", nil)
	req.Header.Set("Authorization", "Bearer sk-bearer-456")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestApiKeyMissing(t *testing.T) {
	r := setupApiKey("sk-test-123")

	req := httptest.NewRequest(http.MethodPost, "/ingest", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestApiKeyInvalid(t *testing.T) {
	r := setupApiKey("sk-test-123")

	req := httptest.NewRequest(http.MethodPost, "/ingest", nil)
	req.Header.Set("X-API-Key", "sk-wrong-key")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestApiKeyBearerCaseInsensitive(t *testing.T) {
	r := setupApiKey("sk-test-789")

	req := httptest.NewRequest(http.MethodPost, "/ingest", nil)
	req.Header.Set("Authorization", "bearer sk-test-789")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestApiKeyBearerMissingKey(t *testing.T) {
	r := setupApiKey("sk-test-123")

	req := httptest.NewRequest(http.MethodPost, "/ingest", nil)
	req.Header.Set("Authorization", "Bearer ")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}
