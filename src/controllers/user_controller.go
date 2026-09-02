package controllers

import (
	"net/http"

	"github.com/CaiqueGOliveira/TelemetryGo/src/application"
	"github.com/CaiqueGOliveira/TelemetryGo/src/application/dtos"
	"github.com/gin-gonic/gin"
)

type UserController struct {
	usecase *application.CreateUserUsecase
}

func NewUserController(usecase *application.CreateUserUsecase) *UserController {
	return &UserController{usecase: usecase}
}

func (uc *UserController) CreateUser(ctx *gin.Context) {
	var dto dtos.UserCreateRequestDto

	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	user, err := uc.usecase.Execute(&dto)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"id":    user.Id,
		"name":  user.Name,
		"email": user.Email.Text(),
	})
}