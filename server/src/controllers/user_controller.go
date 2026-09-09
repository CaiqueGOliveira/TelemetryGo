package controllers

import (
	"net/http"
	"time"

	"github.com/CaiqueGOliveira/TelemetryGo/src/application"
	"github.com/CaiqueGOliveira/TelemetryGo/src/application/dtos"
	t "github.com/CaiqueGOliveira/TelemetryGo/src/application/interfaces"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserController struct {
	createUserUsecase *application.CreateUserUsecase
	loginUsecase      *application.LoginUsecase
	refreshExpiration time.Duration
	jwtProvider       t.TokenProvider
}

func NewUserController(usecase *application.CreateUserUsecase, loginUsecase *application.LoginUsecase, refreshExpiration time.Duration, jwtProvider t.TokenProvider) *UserController {
	return &UserController{
		createUserUsecase: usecase,
		loginUsecase:      loginUsecase,
		refreshExpiration: refreshExpiration,
		jwtProvider:       jwtProvider,
	}
}

func (uc *UserController) CreateUser(ctx *gin.Context) {
	var dto dtos.UserCreateRequestDto

	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	result, err := uc.createUserUsecase.Execute(&dto)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.SetSameSite(http.SameSiteLaxMode)
	ctx.SetCookie(
		"refresh_token",
		result.JwtRefresh,
		int(uc.refreshExpiration.Seconds()),
		"/api/v1/auth/refresh",
		"",
		false,
		true,
	)

	ctx.JSON(http.StatusCreated, gin.H{
		"id":           result.User.Id,
		"name":         result.User.Name,
		"email":        result.User.Email.Text(),
		"access_token": result.JwtAccess,
		"api_key":      result.User.ApiKey,
	})
}

func (uc *UserController) Login(ctx *gin.Context) {
	var dto dtos.LoginRequestDto

	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	result, err := uc.loginUsecase.Execute(dto.Email, dto.Password)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	ctx.SetSameSite(http.SameSiteLaxMode)
	ctx.SetCookie(
		"refresh_token",
		result.JwtRefresh,
		int(uc.refreshExpiration.Seconds()),
		"/api/v1/auth/refresh",
		"",
		false,
		true,
	)

	ctx.JSON(http.StatusOK, gin.H{
		"access_token": result.JwtAccess,
		"api_key":      result.ApiKey,
	})
}

func (uc *UserController) RefreshToken(ctx *gin.Context) {
	raw, err := ctx.Cookie("refresh_token")
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "missing refresh token"})
		return
	}

	claims, err := uc.jwtProvider.VerifyToken(raw)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired refresh token"})
		return
	}

	if typ, _ := claims["typ"].(string); typ != t.TokenTypeRefresh {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token type"})
		return
	}

	sub, ok := claims["sub"].(string)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	userID, err := uuid.Parse(sub)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	newAccess, err := uc.jwtProvider.GenerateToken(userID, t.TokenTypeAccess)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate access token"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"access_token": newAccess,
	})
}

func (uc *UserController) Logout(ctx *gin.Context) {
	ctx.SetSameSite(http.SameSiteLaxMode)
	ctx.SetCookie(
		"refresh_token",
		"",
		-1,
		"/api/v1/auth/refresh",
		"",
		false,
		true,
	)

	ctx.JSON(http.StatusOK, gin.H{
		"message": "logged out",
	})
}
