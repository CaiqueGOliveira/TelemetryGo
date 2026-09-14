package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/CaiqueGOliveira/TelemetryGo/src/application"
	"github.com/CaiqueGOliveira/TelemetryGo/src/infra/auth"
	"github.com/CaiqueGOliveira/TelemetryGo/src/infra/repositories"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type userTestEnv struct {
	router    *gin.Engine
	repo      *repositories.UserRepository
	reqUserID string
}

func newUserTestRouter(t *testing.T) *userTestEnv {
	t.Helper()
	gin.SetMode(gin.TestMode)

	repo := repositories.NewUserRepository()
	jwtSvc, err := auth.NewJwtService("test-secret", time.Hour*24*7, time.Hour/6)
	if err != nil {
		t.Fatalf("failed to init jwt service: %v", err)
	}

	controller := NewUserController(
		application.NewCreateUserUsecase(repo, jwtSvc),
		application.NewLoginUsecase(jwtSvc, repo),
		application.NewGetUserUsecase(repo),
		application.NewUpdateUserUsecase(repo),
		application.NewChangePasswordUsecase(repo),
		application.NewRotateApiKeyUsecase(repo),
		application.NewDeleteUserUsecase(repo),
		application.NewForgotPasswordUsecase(repo, jwtSvc, 30*time.Minute),
		application.NewResetPasswordUsecase(repo, jwtSvc),
		time.Hour*24*7,
		jwtSvc,
	)

	env := &userTestEnv{router: gin.New(), repo: repo}

	env.router.POST("/api/v1/users", controller.CreateUser)
	env.router.POST("/api/v1/login", controller.Login)
	env.router.POST("/api/v1/auth/refresh", controller.RefreshToken)
	env.router.POST("/api/v1/auth/forgot-password", controller.ForgotPassword)
	env.router.POST("/api/v1/auth/reset-password", controller.ResetPassword)

	api := env.router.Group("/api/v1")
	api.Use(func(c *gin.Context) {
		c.Set("user_id", env.reqUserID)
	})
	api.GET("/me", controller.Me)
	api.PUT("/me", controller.UpdateProfile)
	api.DELETE("/users/me", controller.DeleteAccount)
	api.POST("/auth/change-password", controller.ChangePassword)
	api.POST("/auth/rotate-api-key", controller.RotateApiKey)

	return env
}

func createUserAndSetID(t *testing.T, env *userTestEnv) string {
	t.Helper()

	resp := do(t, env.router, http.MethodPost, "/api/v1/users",
		`{"name":"Caique","email":"user@example.com","password":"Senha#Segura1"}`)
	if resp.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", resp.Code, resp.Body.String())
	}

	var body struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("unexpected decode: %v", err)
	}
	env.reqUserID = body.ID
	return body.ID
}

func userCreateBody(name, email, password string) string {
	payload, _ := json.Marshal(map[string]string{"name": name, "email": email, "password": password})
	return string(payload)
}

func TestUserControllerCreate(t *testing.T) {
	env := newUserTestRouter(t)

	resp := do(t, env.router, http.MethodPost, "/api/v1/users",
		userCreateBody("Caique", "user@example.com", "Senha#Segura1"))
	if resp.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", resp.Code, resp.Body.String())
	}

	var body struct {
		ID          string `json:"id"`
		Email       string `json:"email"`
		ApiKey      string `json:"api_key"`
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("unexpected decode: %v", err)
	}
	if body.ID == "" || body.ApiKey == "" || body.AccessToken == "" {
		t.Fatalf("expected full payload, got %+v", body)
	}
	if body.Email != "user@example.com" {
		t.Errorf("unexpected email %s", body.Email)
	}

	cookies := resp.Result().Cookies()
	found := false
	for _, cookie := range cookies {
		if cookie.Name == "refresh_token" {
			found = true
		}
	}
	if !found {
		t.Error("expected refresh_token cookie")
	}
}

func TestUserControllerCreateDuplicateEmail(t *testing.T) {
	env := newUserTestRouter(t)

	first := do(t, env.router, http.MethodPost, "/api/v1/users",
		userCreateBody("Caique", "user@example.com", "Senha#Segura1"))
	if first.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", first.Code)
	}

	second := do(t, env.router, http.MethodPost, "/api/v1/users",
		userCreateBody("Outro", "user@example.com", "Senha#Segura1"))
	if second.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", second.Code)
	}
}

func TestUserControllerCreateInvalidBody(t *testing.T) {
	env := newUserTestRouter(t)

	resp := do(t, env.router, http.MethodPost, "/api/v1/users", `not json`)
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.Code)
	}
}

func TestUserControllerLogin(t *testing.T) {
	env := newUserTestRouter(t)

	do(t, env.router, http.MethodPost, "/api/v1/users",
		userCreateBody("Caique", "user@example.com", "Senha#Segura1"))

	resp := do(t, env.router, http.MethodPost, "/api/v1/login",
		`{"email":"user@example.com","password":"Senha#Segura1"}`)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.Code, resp.Body.String())
	}

	var body struct {
		AccessToken string `json:"access_token"`
		ApiKey      string `json:"api_key"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("unexpected decode: %v", err)
	}
	if body.AccessToken == "" || body.ApiKey == "" {
		t.Error("expected tokens in login response")
	}
}

func TestUserControllerLoginWrongCredentials(t *testing.T) {
	env := newUserTestRouter(t)

	resp := do(t, env.router, http.MethodPost, "/api/v1/login",
		`{"email":"missing@example.com","password":"Senha#Segura1"}`)
	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.Code)
	}
}

func TestUserControllerMe(t *testing.T) {
	env := newUserTestRouter(t)
	createdID := createUserAndSetID(t, env)

	resp := do(t, env.router, http.MethodGet, "/api/v1/me", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.Code, resp.Body.String())
	}

	var body struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Email  string `json:"email"`
		ApiKey string `json:"api_key"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("unexpected decode: %v", err)
	}
	if body.ID != createdID {
		t.Errorf("expected id %s, got %s", createdID, body.ID)
	}
	if body.Name != "Caique" || body.Email != "user@example.com" {
		t.Errorf("unexpected profile: %+v", body)
	}
}

func TestUserControllerMeNotFound(t *testing.T) {
	env := newUserTestRouter(t)
	env.reqUserID = uuid.New().String()

	resp := do(t, env.router, http.MethodGet, "/api/v1/me", "")
	if resp.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.Code)
	}
}

func TestUserControllerUpdateProfile(t *testing.T) {
	env := newUserTestRouter(t)
	createUserAndSetID(t, env)

	resp := do(t, env.router, http.MethodPut, "/api/v1/me",
		`{"name":"Novo Nome","email":"novo@example.com"}`)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.Code, resp.Body.String())
	}

	var body struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("unexpected decode: %v", err)
	}
	if body.Name != "Novo Nome" || body.Email != "novo@example.com" {
		t.Errorf("unexpected profile after update: %+v", body)
	}
}

func TestUserControllerUpdateProfileNotFound(t *testing.T) {
	env := newUserTestRouter(t)
	env.reqUserID = uuid.New().String()

	resp := do(t, env.router, http.MethodPut, "/api/v1/me",
		`{"name":"X","email":"x@example.com"}`)
	if resp.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.Code)
	}
}

func TestUserControllerChangePassword(t *testing.T) {
	env := newUserTestRouter(t)
	createUserAndSetID(t, env)

	ok := do(t, env.router, http.MethodPost, "/api/v1/auth/change-password",
		`{"current_password":"Senha#Segura1","new_password":"NovaSenha#1"}`)
	if ok.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", ok.Code, ok.Body.String())
	}

	wrong := do(t, env.router, http.MethodPost, "/api/v1/auth/change-password",
		`{"current_password":"errado#1","new_password":"NovaSenha#2"}`)
	if wrong.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 on wrong current password, got %d", wrong.Code)
	}
}

func TestUserControllerRotateApiKey(t *testing.T) {
	env := newUserTestRouter(t)
	createdID := createUserAndSetID(t, env)

	resp := do(t, env.router, http.MethodPost, "/api/v1/auth/rotate-api-key", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.Code, resp.Body.String())
	}

	var body struct {
		ApiKey string `json:"api_key"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("unexpected decode: %v", err)
	}
	if body.ApiKey == "" {
		t.Fatal("expected api_key in response")
	}

	user, err := env.repo.FindById(uuid.MustParse(createdID))
	if err != nil {
		t.Fatal(err)
	}
	if user.ApiKey != body.ApiKey {
		t.Error("expected stored api key to be rotated")
	}
}

func TestUserControllerDeleteAccount(t *testing.T) {
	env := newUserTestRouter(t)
	createUserAndSetID(t, env)

	resp := do(t, env.router, http.MethodDelete, "/api/v1/users/me", "")
	if resp.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.Code)
	}
}

func TestUserControllerDeleteAccountNotFound(t *testing.T) {
	env := newUserTestRouter(t)
	env.reqUserID = uuid.New().String()

	resp := do(t, env.router, http.MethodDelete, "/api/v1/users/me", "")
	if resp.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.Code)
	}
}

func TestUserControllerRefreshWithoutCookie(t *testing.T) {
	env := newUserTestRouter(t)

	resp := do(t, env.router, http.MethodPost, "/api/v1/auth/refresh", "")
	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.Code)
	}
}

func TestUserControllerForgotPasswordExistingEmail(t *testing.T) {
	env := newUserTestRouter(t)
	createUserAndSetID(t, env)

	resp := do(t, env.router, http.MethodPost, "/api/v1/auth/forgot-password",
		`{"email":"user@example.com"}`)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.Code, resp.Body.String())
	}

	var body map[string]interface{}
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("unexpected decode: %v", err)
	}
	if body["reset_token"] == "" {
		t.Fatal("expected reset_token in response for existing email")
	}
}

func TestUserControllerForgotPasswordMissingEmail(t *testing.T) {
	env := newUserTestRouter(t)

	resp := do(t, env.router, http.MethodPost, "/api/v1/auth/forgot-password",
		`{"email":"missing@example.com"}`)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.Code, resp.Body.String())
	}

	var body map[string]interface{}
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("unexpected decode: %v", err)
	}
	if _, ok := body["reset_token"]; ok {
		t.Fatal("expected no reset_token for missing email")
	}
}

func TestUserControllerForgotPasswordInvalidBody(t *testing.T) {
	env := newUserTestRouter(t)

	resp := do(t, env.router, http.MethodPost, "/api/v1/auth/forgot-password",
		`{"email":"not-an-email"}`)
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.Code)
	}
}

func TestUserControllerResetPasswordSuccess(t *testing.T) {
	env := newUserTestRouter(t)
	createUserAndSetID(t, env)

	forgot := do(t, env.router, http.MethodPost, "/api/v1/auth/forgot-password",
		`{"email":"user@example.com"}`)
	var forgotBody map[string]interface{}
	if err := json.Unmarshal(forgot.Body.Bytes(), &forgotBody); err != nil {
		t.Fatalf("unexpected decode: %v", err)
	}
	token, _ := forgotBody["reset_token"].(string)
	if token == "" {
		t.Fatal("expected reset token from forgot-password")
	}

	reset := do(t, env.router, http.MethodPost, "/api/v1/auth/reset-password",
		fmt.Sprintf(`{"token":"%s","new_password":"NovaSenha#1"}`, token))
	if reset.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", reset.Code, reset.Body.String())
	}

	login := do(t, env.router, http.MethodPost, "/api/v1/login",
		`{"email":"user@example.com","password":"NovaSenha#1"}`)
	if login.Code != http.StatusOK {
		t.Fatalf("expected login with new password to succeed, got %d", login.Code)
	}
}

func TestUserControllerResetPasswordInvalidToken(t *testing.T) {
	env := newUserTestRouter(t)

	resp := do(t, env.router, http.MethodPost, "/api/v1/auth/reset-password",
		`{"token":"invalid","new_password":"NovaSenha#1"}`)
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.Code)
	}
}

func TestUserControllerResetPasswordInvalidBody(t *testing.T) {
	env := newUserTestRouter(t)

	resp := do(t, env.router, http.MethodPost, "/api/v1/auth/reset-password",
		`{"token":"valid"}`)
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.Code)
	}
}
