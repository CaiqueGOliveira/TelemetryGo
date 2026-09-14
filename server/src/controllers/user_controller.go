package controllers

import (
	"errors"
	"net/http"
	"time"

	"github.com/CaiqueGOliveira/TelemetryGo/src/application"
	"github.com/CaiqueGOliveira/TelemetryGo/src/application/dtos"
	t "github.com/CaiqueGOliveira/TelemetryGo/src/application/interfaces"
	r "github.com/CaiqueGOliveira/TelemetryGo/src/domain/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserController struct {
	createUserUsecase     *application.CreateUserUsecase
	loginUsecase          *application.LoginUsecase
	getUserUsecase        *application.GetUserUsecase
	updateUserUsecase     *application.UpdateUserUsecase
	changePasswordUsecase *application.ChangePasswordUsecase
	rotateApiKeyUsecase   *application.RotateApiKeyUsecase
	deleteUserUsecase     *application.DeleteUserUsecase
	forgotPasswordUsecase *application.ForgotPasswordUsecase
	resetPasswordUsecase  *application.ResetPasswordUsecase
	refreshExpiration     time.Duration
	jwtProvider           t.TokenProvider
}

func NewUserController(
	createUserUsecase *application.CreateUserUsecase,
	loginUsecase *application.LoginUsecase,
	getUserUsecase *application.GetUserUsecase,
	updateUserUsecase *application.UpdateUserUsecase,
	changePasswordUsecase *application.ChangePasswordUsecase,
	rotateApiKeyUsecase *application.RotateApiKeyUsecase,
	deleteUserUsecase *application.DeleteUserUsecase,
	forgotPasswordUsecase *application.ForgotPasswordUsecase,
	resetPasswordUsecase *application.ResetPasswordUsecase,
	refreshExpiration time.Duration,
	jwtProvider t.TokenProvider,
) *UserController {
	return &UserController{
		createUserUsecase:     createUserUsecase,
		loginUsecase:          loginUsecase,
		getUserUsecase:        getUserUsecase,
		updateUserUsecase:     updateUserUsecase,
		changePasswordUsecase: changePasswordUsecase,
		rotateApiKeyUsecase:   rotateApiKeyUsecase,
		deleteUserUsecase:     deleteUserUsecase,
		forgotPasswordUsecase: forgotPasswordUsecase,
		resetPasswordUsecase:  resetPasswordUsecase,
		refreshExpiration:     refreshExpiration,
		jwtProvider:           jwtProvider,
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
		if errors.Is(err, r.ErrDuplicateEmail) {
			ctx.JSON(http.StatusConflict, gin.H{"error": "email already in use"})
			return
		}
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

func (uc *UserController) Me(ctx *gin.Context) {
	result, err := uc.getUserUsecase.Execute(ctx.GetString("user_id"))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	ctx.JSON(http.StatusOK, result)
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

func (uc *UserController) UpdateProfile(ctx *gin.Context) {
	var dto dtos.UpdateUserRequestDto

	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	result, err := uc.updateUserUsecase.Execute(ctx.GetString("user_id"), &dto)
	if err != nil {
		switch {
		case errors.Is(err, r.ErrDuplicateEmail):
			ctx.JSON(http.StatusConflict, gin.H{"error": "email already in use"})
			return
		case errors.Is(err, r.ErrNotFound):
			ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		default:
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	ctx.JSON(http.StatusOK, result)
}

func (uc *UserController) ChangePassword(ctx *gin.Context) {
	var dto dtos.ChangePasswordRequestDto

	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	err := uc.changePasswordUsecase.Execute(ctx.GetString("user_id"), dto.CurrentPassword, dto.NewPassword)
	if err != nil {
		if errors.Is(err, r.ErrNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "password changed"})
}

func (uc *UserController) RotateApiKey(ctx *gin.Context) {
	apiKey, err := uc.rotateApiKeyUsecase.Execute(ctx.GetString("user_id"))
	if err != nil {
		if errors.Is(err, r.ErrNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"api_key": apiKey})
}

func (uc *UserController) DeleteAccount(ctx *gin.Context) {
	err := uc.deleteUserUsecase.Execute(ctx.GetString("user_id"))
	if err != nil {
		if errors.Is(err, r.ErrNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (uc *UserController) ForgotPassword(ctx *gin.Context) {
	var dto dtos.ForgotPasswordRequestDto

	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	resetToken, err := uc.forgotPasswordUsecase.Execute(dto.Email)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to request password reset"})
		return
	}

	body := gin.H{
		"message": "if the email exists, password reset instructions were sent",
	}

	// Demo: sem infraestrutura de email, o token é devolvido na resposta.
	// Em produção, ele deve ser enviado por email e nunca retornado aqui.
	if resetToken != "" {
		body["reset_token"] = resetToken
	}

	ctx.JSON(http.StatusOK, body)
}

func (uc *UserController) ResetPassword(ctx *gin.Context) {
	var dto dtos.ResetPasswordRequestDto

	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	err := uc.resetPasswordUsecase.Execute(dto.Token, dto.NewPassword)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "password has been reset"})
}
