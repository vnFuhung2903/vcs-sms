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
	authRoutes := r.Group("/users")
	{
		authRoutes.POST("/register", h.Register)
		authRoutes.POST("/login", h.Login)
		authRoutes.PUT("/update/password/:id", h.UpdatePassword)
	}
}

// Register godoc
// @Summary Register a new user
// @Description Register a user and return a JWT token
// @Tags users
// @Accept json
// @Produce json
// @Param body body dto.RegisterRequest true "User registration request"
// @Success 200 {object} dto.MessageResponse "User registered successfully"
// @Failure 400 {object} dto.APIResponse "Bad request"
// @Failure 500 {object} dto.APIResponse "Internal server error"
// @Router /users/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "BAD_REQUEST",
			Message: "Invalid request data",
			Error:   err,
		})
		return
	}

	scopes := utils.UserRoleToDefaultScopes(req.Role, nil)
	_, err := h.authService.Register(req.Username, req.Password, req.Email, req.Role, utils.ScopesToHashMap(scopes))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Error: err.Error(),
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
// @Tags users
// @Accept json
// @Produce json
// @Param body body dto.LoginRequest true "User login credentials"
// @Success 200 {object} dto.MessageResponse "Login successful"
// @Failure 400 {object} dto.APIResponse "Bad request"
// @Failure 401 {object} dto.APIResponse "Unauthorized"
// @Failure 500 {object} dto.APIResponse "Internal server error"
// @Router /users/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "BAD_REQUEST",
			Message: "Invalid request data",
			Error:   err,
		})
		return
	}

	accessToken, err := h.authService.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.APIResponse{
			Error: err.Error(),
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
// @Description Update password of currently logged-in user
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param body body dto.UpdatePasswordRequest true "New password"
// @Success 200 {object} dto.MessageResponse "Password updated successfully"
// @Failure 400 {object} dto.APIResponse "Bad request"
// @Failure 500 {object} dto.APIResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /users/update/password/{id} [put]
func (h *AuthHandler) UpdatePassword(c *gin.Context) {
	userId := c.GetString("userId")
	var req dto.UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "BAD_REQUEST",
			Message: "Invalid request data",
			Error:   err,
		})
		return
	}

	if err := h.authService.UpdatePassword(userId, req.CurrentPassword, req.NewPassword); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to update password",
			Error:   err,
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "UPDATE_SUCCESS",
		Message: "Password updated successfully",
	})
}
