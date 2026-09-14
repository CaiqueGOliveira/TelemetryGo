package integration

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/CaiqueGOliveira/TelemetryGo/src/application"
	itf "github.com/CaiqueGOliveira/TelemetryGo/src/application/interfaces"
	"github.com/CaiqueGOliveira/TelemetryGo/src/controllers"
	"github.com/CaiqueGOliveira/TelemetryGo/src/domain/messages"
	"github.com/CaiqueGOliveira/TelemetryGo/src/infra/auth"
	"github.com/CaiqueGOliveira/TelemetryGo/src/infra/messaging"
	"github.com/CaiqueGOliveira/TelemetryGo/src/infra/middleware"
	"github.com/CaiqueGOliveira/TelemetryGo/src/infra/repositories"
	"github.com/CaiqueGOliveira/TelemetryGo/src/infra/routes"
	"github.com/gin-gonic/gin"
)

func setupRouter(t *testing.T) *gin.Engine {
	return setupRouterWith(t, messaging.NewNoOpPublisher())
}

func setupRouterWith(t *testing.T, eventPublisher messages.EventPublisher) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	repo := repositories.NewUserRepository()

	var tokenProvider itf.TokenProvider
	tokenProvider, err := auth.NewJwtService("test-secret", time.Hour*24*7, time.Hour/6)
	if err != nil {
		log.Fatalf("failed to initialize jwt service: %v", err)
	}

	createUserUsecase := application.NewCreateUserUsecase(repo, tokenProvider)
	loginUsecase := application.NewLoginUsecase(tokenProvider, repo)
	getUserUsecase := application.NewGetUserUsecase(repo)
	updateUserUsecase := application.NewUpdateUserUsecase(repo)
	changePasswordUsecase := application.NewChangePasswordUsecase(repo)
	rotateApiKeyUsecase := application.NewRotateApiKeyUsecase(repo)
	deleteUserUsecase := application.NewDeleteUserUsecase(repo)
	forgotPasswordUsecase := application.NewForgotPasswordUsecase(repo, tokenProvider, 30*time.Minute)
	resetPasswordUsecase := application.NewResetPasswordUsecase(repo, tokenProvider)

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
		time.Hour*24*7,
		tokenProvider,
	)

	eventRepo := repositories.NewInMemoryEventRepository()
	eventUsecase := application.NewEventUsecase(eventRepo, eventPublisher)
	eventController := controllers.NewEventController(eventUsecase)

	metricRepo := repositories.NewInMemoryMetricRepository()
	metricUsecase := application.NewMetricUsecase(metricRepo, eventPublisher)
	metricController := controllers.NewMetricController(metricUsecase)

	highLimit := middleware.RateLimitConfig{RPS: 100000, Burst: 100000}

	return routes.SetupRouter(userController, eventController, metricController, tokenProvider, repo, highLimit, highLimit, nil)
}

func createUserAndGetApiKeyWithEmail(t *testing.T, r *gin.Engine, email string) string {
	t.Helper()

	resp := doJSON(t, r, http.MethodPost, "/api/v1/users", map[string]string{
		"name":     "Caique",
		"email":    email,
		"password": "Senha#Segura1",
	}, "")
	if resp.Code != http.StatusCreated {
		t.Fatalf("expected 201 on create, got %d: %s", resp.Code, resp.Body.String())
	}

	var body struct {
		ApiKey string `json:"api_key"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("unexpected error decoding create body: %v", err)
	}
	if body.ApiKey == "" {
		t.Fatal("expected api_key in create response")
	}

	return body.ApiKey
}

func createUserAndGetApiKey(t *testing.T, r *gin.Engine) string {
	return createUserAndGetApiKeyWithEmail(t, r, "user@example.com")
}

func loginAndGetAccessToken(t *testing.T, r *gin.Engine, email string) string {
	t.Helper()

	resp := doJSON(t, r, http.MethodPost, "/api/v1/login", map[string]string{
		"email":    email,
		"password": "Senha#Segura1",
	}, "")
	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200 on login, got %d: %s", resp.Code, resp.Body.String())
	}

	var body struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("unexpected error decoding login body: %v", err)
	}
	if body.AccessToken == "" {
		t.Fatal("expected access_token in login response")
	}

	return body.AccessToken
}

func doJSONWithCookies(t *testing.T, r *gin.Engine, method, path string, body interface{}, cookies []*http.Cookie) *httptest.ResponseRecorder {
	t.Helper()

	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("unexpected error encoding body: %v", err)
		}
	}

	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func findRefreshCookie(t *testing.T, w *httptest.ResponseRecorder) *http.Cookie {
	t.Helper()

	for _, cookie := range w.Result().Cookies() {
		if cookie.Name == "refresh_token" {
			return cookie
		}
	}
	return nil
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

	var createBody struct {
		ID          string `json:"id"`
		ApiKey      string `json:"api_key"`
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(createResp.Body.Bytes(), &createBody); err != nil {
		t.Fatalf("unexpected error decoding create body: %v", err)
	}
	if createBody.ApiKey == "" {
		t.Fatal("expected api_key in create response")
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
		ApiKey      string `json:"api_key"`
	}
	if err := json.Unmarshal(loginResp.Body.Bytes(), &loginBody); err != nil {
		t.Fatalf("unexpected error decoding login body: %v", err)
	}
	if loginBody.AccessToken == "" {
		t.Fatal("expected access_token in login response")
	}
	if loginBody.ApiKey == "" {
		t.Fatal("expected api_key in login response")
	}
	if loginBody.ApiKey != createBody.ApiKey {
		t.Errorf("expected api key from login to match create, got %s != %s", loginBody.ApiKey, createBody.ApiKey)
	}

	meResp := doJSON(t, r, http.MethodGet, "/api/v1/me", nil, loginBody.AccessToken)
	if meResp.Code != http.StatusOK {
		t.Fatalf("expected 200 on /me, got %d: %s", meResp.Code, meResp.Body.String())
	}

	var meBody struct {
		Id     string `json:"id"`
		Name   string `json:"name"`
		Email  string `json:"email"`
		ApiKey string `json:"api_key"`
	}
	if err := json.Unmarshal(meResp.Body.Bytes(), &meBody); err != nil {
		t.Fatalf("unexpected error decoding me body: %v", err)
	}
	if meBody.Name != "Caique" {
		t.Errorf("expected /me name Caique, got %q", meBody.Name)
	}
	if meBody.Email != "user@example.com" {
		t.Errorf("expected /me email user@example.com, got %q", meBody.Email)
	}
	if meBody.ApiKey != createBody.ApiKey {
		t.Errorf("expected /me api_key to match create, got %q != %q", meBody.ApiKey, createBody.ApiKey)
	}
	if meBody.Id == "" {
		t.Error("expected /me to return id")
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

func TestRefreshTokenFlow(t *testing.T) {
	r := setupRouter(t)

	createResp := doJSON(t, r, http.MethodPost, "/api/v1/users", map[string]string{
		"name":     "Caique",
		"email":    "user@example.com",
		"password": "Senha#Segura1",
	}, "")
	if createResp.Code != http.StatusCreated {
		t.Fatalf("expected 201 on create, got %d: %s", createResp.Code, createResp.Body.String())
	}

	refreshCookie := findRefreshCookie(t, createResp)
	if refreshCookie == nil {
		t.Fatal("expected refresh_token cookie on create")
	}

	refreshResp := doJSONWithCookies(t, r, http.MethodPost, "/api/v1/auth/refresh", nil, []*http.Cookie{refreshCookie})
	if refreshResp.Code != http.StatusOK {
		t.Fatalf("expected 200 on refresh, got %d: %s", refreshResp.Code, refreshResp.Body.String())
	}

	var body struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(refreshResp.Body.Bytes(), &body); err != nil {
		t.Fatalf("unexpected error decoding refresh body: %v", err)
	}
	if body.AccessToken == "" {
		t.Fatal("expected access_token in refresh response")
	}
}

func TestRefreshRejectsAccessToken(t *testing.T) {
	r := setupRouter(t)

	createResp := doJSON(t, r, http.MethodPost, "/api/v1/users", map[string]string{
		"name":     "Caique",
		"email":    "user@example.com",
		"password": "Senha#Segura1",
	}, "")
	if createResp.Code != http.StatusCreated {
		t.Fatalf("expected 201 on create, got %d: %s", createResp.Code, createResp.Body.String())
	}

	accessToken := loginAndGetAccessToken(t, r, "user@example.com")

	accessCookie := &http.Cookie{Name: "refresh_token", Value: accessToken}
	refreshResp := doJSONWithCookies(t, r, http.MethodPost, "/api/v1/auth/refresh", nil, []*http.Cookie{accessCookie})
	if refreshResp.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 when using access token to refresh, got %d", refreshResp.Code)
	}
}

func TestRefreshWithoutCookie(t *testing.T) {
	r := setupRouter(t)

	resp := doJSON(t, r, http.MethodPost, "/api/v1/auth/refresh", nil, "")
	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without cookie, got %d", resp.Code)
	}
}

func TestCreateUserDuplicateEmail(t *testing.T) {
	r := setupRouter(t)

	first := doJSON(t, r, http.MethodPost, "/api/v1/users", map[string]string{
		"name":     "Caique",
		"email":    "user@example.com",
		"password": "Senha#Segura1",
	}, "")
	if first.Code != http.StatusCreated {
		t.Fatalf("expected 201 on create, got %d: %s", first.Code, first.Body.String())
	}

	duplicate := doJSON(t, r, http.MethodPost, "/api/v1/users", map[string]string{
		"name":     "Outro",
		"email":    "user@example.com",
		"password": "Senha#Segura1",
	}, "")
	if duplicate.Code != http.StatusConflict {
		t.Fatalf("expected 409 on duplicate email, got %d: %s", duplicate.Code, duplicate.Body.String())
	}
	if !bytes.Contains(duplicate.Body.Bytes(), []byte("email already in use")) {
		t.Fatalf("expected error message about duplicate email, got: %s", duplicate.Body.String())
	}
}

func TestLogoutClearsRefreshCookie(t *testing.T) {
	r := setupRouter(t)

	createResp := doJSON(t, r, http.MethodPost, "/api/v1/users", map[string]string{
		"name":     "Caique",
		"email":    "user@example.com",
		"password": "Senha#Segura1",
	}, "")
	if createResp.Code != http.StatusCreated {
		t.Fatalf("expected 201 on create, got %d: %s", createResp.Code, createResp.Body.String())
	}

	refreshCookie := findRefreshCookie(t, createResp)
	if refreshCookie == nil {
		t.Fatal("expected refresh_token cookie on create")
	}

	logoutResp := doJSONWithCookies(t, r, http.MethodPost, "/api/v1/auth/logout", nil, []*http.Cookie{refreshCookie})
	if logoutResp.Code != http.StatusOK {
		t.Fatalf("expected 200 on logout, got %d", logoutResp.Code)
	}

	cleared := findRefreshCookie(t, logoutResp)
	if cleared == nil {
		t.Fatal("expected a set-cookie clearing refresh_token on logout")
	}
	if cleared.MaxAge >= 0 && !cleared.Expires.Before(time.Now()) {
		t.Fatal("expected logout cookie to be expired")
	}
}
