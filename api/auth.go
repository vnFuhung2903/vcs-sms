package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vnFuhung2903/vcs-sms/dto"
	"github.com/vnFuhung2903/vcs-sms/pkg/middlewares"
	"github.com/vnFuhung2903/vcs-sms/usecases/services"
	"github.com/vnFuhung2903/vcs-sms/utils"
)

type AuthHandler struct {
	authService   services.IAuthService
	jwtMiddleware middlewares.IJWTMiddleware
}

func NewAuthHandler(authService services.IAuthService, jwtMiddleware middlewares.IJWTMiddleware) *AuthHandler {
	return &AuthHandler{authService, jwtMiddleware}
}

func (h *AuthHandler) SetupRoutes(r *gin.Engine) {
	authRoutes := r.Group("/auth")
	{
		authRoutes.POST("/register", h.Register)
		authRoutes.POST("/login", h.Login)

		authRequiredGroup := authRoutes.Group("", h.jwtMiddleware.RequireScope(""))
		{
			authRequiredGroup.PUT("/update/password", h.UpdatePassword)
			authRequiredGroup.POST("/refresh", h.RefreshAccessToken)
		}
	}
}

// Register godoc
// @Summary Register a new user
// @Description Register a user and return a JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param body body dto.RegisterRequest true "User registration request"
// @Success 201 {object} dto.APIResponse "User registered successfully"
// @Failure 400 {object} dto.APIResponse "Bad request"
// @Failure 500 {object} dto.APIResponse "Internal server error"
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "BAD_REQUEST",
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	scopes := utils.UserRoleToDefaultScopes(req.Role, nil)
	_, err := h.authService.Register(req.Username, req.Password, req.Email, req.Role, utils.ScopesToHashMap(scopes))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to register user",
			Error:   err.Error(),
		})
		return
	}
	c.JSON(http.StatusCreated, dto.APIResponse{
		Success: true,
		Code:    "REGISTER_SUCCESS",
		Message: "User registered successfully",
	})
}

// Login godoc
// @Summary Login with username and password
// @Description Login and receive JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param body body dto.LoginRequest true "User login credentials"
// @Success 200 {object} dto.APIResponse "Login successful"
// @Failure 400 {object} dto.APIResponse "Bad request"
// @Failure 500 {object} dto.APIResponse "Internal server error"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "BAD_REQUEST",
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	accessToken, err := h.authService.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to login",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "LOGIN_SUCCESS",
		Message: "Login successful",
		Data: dto.LoginResponse{
			AccessToken: accessToken,
		},
	})
}

// UpdatePassword godoc
// @Summary Update own password
// @Description Update the password of the currently authenticated user
// @Tags auth
// @Accept json
// @Produce json
// @Param body body dto.UpdatePasswordRequest true "New password request"
// @Success 200 {object} dto.APIResponse "Password updated successfully"
// @Failure 400 {object} dto.APIResponse "Bad request"
// @Failure 500 {object} dto.APIResponse "Internal server error"
// @Security BearerAuth
// @Router /auth/update/password [put]
func (h *AuthHandler) UpdatePassword(c *gin.Context) {
	userId := c.GetString("userId")
	var req dto.UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "BAD_REQUEST",
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	if err := h.authService.UpdatePassword(c.Request.Context(), userId, req.CurrentPassword, req.NewPassword); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to update password",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "PASSWORD_UPDATE_SUCCESS",
		Message: "Password updated successfully",
	})
}

// RefreshAccessToken godoc
// @Summary Refresh access token
// @Description Refresh the access token for the currently authenticated user
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} dto.APIResponse "Access token refreshed successfully"
// @Failure 500 {object} dto.APIResponse "Internal server error"
// @Security BearerAuth
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshAccessToken(c *gin.Context) {
	userId := c.GetString("userId")
	accessToken, err := h.authService.RefreshAccessToken(c.Request.Context(), userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to update password",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "REFRESH_SUCCESS",
		Message: "Access token refreshed successfully",
		Data: dto.LoginResponse{
			AccessToken: accessToken,
		},
	})
}
