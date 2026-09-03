package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/CaiqueGOliveira/TelemetryGo/src/application"
	itf "github.com/CaiqueGOliveira/TelemetryGo/src/application/interfaces"
	"github.com/CaiqueGOliveira/TelemetryGo/src/controllers"
	"github.com/CaiqueGOliveira/TelemetryGo/src/infra/auth"
	"github.com/CaiqueGOliveira/TelemetryGo/src/infra/repositories"
	"github.com/CaiqueGOliveira/TelemetryGo/src/infra/routes"
	"github.com/gin-gonic/gin"
)

func setupRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	repo := repositories.NewUserRepository()

	var tokenProvider itf.TokenProvider
	tokenProvider = auth.NewJwtService("test-secret", time.Hour*24*7, time.Hour/6)

	createUserUsecase := application.NewCreateUserUsecase(repo, tokenProvider)
	loginUsecase := application.NewLoginUsecase(tokenProvider, repo)

	userController := controllers.NewUserController(createUserUsecase, loginUsecase, time.Hour*24*7)

	return routes.SetupRouter(userController, tokenProvider)
}

func doJSON(t *testing.T, r *gin.Engine, method, path string, body interface{}, token string) *httptest.ResponseRecorder {
	t.Helper()

	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("unexpected error encoding body: %v", err)
		}
	}

	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestFullAuthFlow(t *testing.T) {
	r := setupRouter(t)

	createResp := doJSON(t, r, http.MethodPost, "/api/v1/users", map[string]string{
		"name":     "Caique",
		"email":    "user@example.com",
		"password": "Senha#Segura1",
	}, "")
	if createResp.Code != http.StatusCreated {
		t.Fatalf("expected 201 on create, got %d: %s", createResp.Code, createResp.Body.String())
	}

	loginResp := doJSON(t, r, http.MethodPost, "/api/v1/login", map[string]string{
		"email":    "user@example.com",
		"password": "Senha#Segura1",
	}, "")
	if loginResp.Code != http.StatusOK {
		t.Fatalf("expected 200 on login, got %d: %s", loginResp.Code, loginResp.Body.String())
	}

	var loginBody struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(loginResp.Body.Bytes(), &loginBody); err != nil {
		t.Fatalf("unexpected error decoding login body: %v", err)
	}
	if loginBody.AccessToken == "" {
		t.Fatal("expected access_token in login response")
	}

	meResp := doJSON(t, r, http.MethodGet, "/api/v1/me", nil, loginBody.AccessToken)
	if meResp.Code != http.StatusOK {
		t.Fatalf("expected 200 on /me, got %d: %s", meResp.Code, meResp.Body.String())
	}
}

func TestLoginEndpointWrongCredentials(t *testing.T) {
	r := setupRouter(t)

	resp := doJSON(t, r, http.MethodPost, "/api/v1/login", map[string]string{
		"email":    "missing@example.com",
		"password": "Senha#Segura1",
	}, "")
	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 on bad login, got %d", resp.Code)
	}
}

func TestMeEndpointWithoutToken(t *testing.T) {
	r := setupRouter(t)

	resp := doJSON(t, r, http.MethodGet, "/api/v1/me", nil, "")
	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without token, got %d", resp.Code)
	}
}
