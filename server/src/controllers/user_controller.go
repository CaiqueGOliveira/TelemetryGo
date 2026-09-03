package controllers

import (
	"net/http"
	"time"

	"github.com/CaiqueGOliveira/TelemetryGo/src/application"
	"github.com/CaiqueGOliveira/TelemetryGo/src/application/dtos"
	"github.com/gin-gonic/gin"
)

type UserController struct {
	createUserUsecase *application.CreateUserUsecase
	loginUsecase      *application.LoginUsecase
	refreshExpiration time.Duration
}

func NewUserController(usecase *application.CreateUserUsecase, loginUsecase *application.LoginUsecase, refreshExpiration time.Duration) *UserController {
	return &UserController{
		createUserUsecase: usecase,
		loginUsecase:      loginUsecase,
		refreshExpiration: refreshExpiration,
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

	ctx.SetCookie("refresh_token", result.JwtRefresh, int(uc.refreshExpiration.Seconds()), "/", "", false, true)

	ctx.JSON(http.StatusCreated, gin.H{
		"id":           result.User.Id,
		"name":         result.User.Name,
		"email":        result.User.Email.Text(),
		"access_token": result.JwtAccess,
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

	ctx.SetCookie("refresh_token", result.JwtRefresh, int(uc.refreshExpiration.Seconds()), "/", "", false, true)

	ctx.JSON(http.StatusOK, gin.H{
		"access_token": result.JwtAccess,
	})
}
