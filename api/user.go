package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vnFuhung2903/vcs-sms/dto"
	"github.com/vnFuhung2903/vcs-sms/pkg/middlewares"
	"github.com/vnFuhung2903/vcs-sms/usecases/services"
	"github.com/vnFuhung2903/vcs-sms/utils"
)

type UserHandler struct {
	userService   services.IUserService
	jwtMiddleware middlewares.IJWTMiddleware
}

func NewUserHandler(userService services.IUserService, jwtMiddleware middlewares.IJWTMiddleware) *UserHandler {
	return &UserHandler{userService, jwtMiddleware}
}

func (h *UserHandler) SetupRoutes(r *gin.Engine) {
	userRoutes := r.Group("/users")
	{
		userRoutes.POST("/register", h.Register)
		userRoutes.POST("/login", h.Login)

		userGroup := userRoutes.Group("", h.jwtMiddleware.RequireScope("user:modify"))
		{
			userGroup.PUT("/update/password/:id", h.UpdatePassword)
		}

		adminGroup := userRoutes.Group("", h.jwtMiddleware.RequireScope("user:manage"))
		{
			adminGroup.PUT("/update/role/:id", h.UpdateRole)
			adminGroup.PUT("/update/scope/:id", h.UpdateScope)
			adminGroup.DELETE("/delete/:id", h.Delete)
		}
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
// @Failure 400 {object} dto.ErrorResponse "Bad request"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /users/register [post]
func (h *UserHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	scopes := utils.UserRoleToDefaultScopes(req.Role, nil)
	user, err := h.userService.Register(req.Username, req.Password, req.Email, req.Role, utils.ScopesToHashMap(scopes))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	if err := h.jwtMiddleware.GenerateJWT(c.Request.Context(), user.ID, user.Username, utils.HashMapToScopes(user.Scopes)); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, dto.MessageResponse{
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
// @Failure 400 {object} dto.ErrorResponse "Bad request"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /users/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	user, err := h.userService.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	if err := h.jwtMiddleware.GenerateJWT(c.Request.Context(), user.ID, user.Username, utils.HashMapToScopes(user.Scopes)); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "Login successful",
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
// @Failure 400 {object} dto.ErrorResponse "Bad request"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /users/update/password/{id} [put]
func (h *UserHandler) UpdatePassword(c *gin.Context) {
	userId := c.GetString("userId")
	var req dto.UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	if err := h.userService.UpdatePassword(userId, req.Password); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "Password updated successfully",
	})
}

// UpdateRole godoc
// @Summary Update a user's role
// @Description Update role of a user (admin only)
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param body body dto.UpdateRoleRequest true "New role"
// @Success 200 {object} dto.MessageResponse "Role updated successfully"
// @Failure 400 {object} dto.ErrorResponse "Bad request"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /users/update/role/{id} [put]
func (h *UserHandler) UpdateRole(c *gin.Context) {
	userId := c.GetString("userId")
	var req dto.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	if err := h.userService.UpdateRole(userId, req.Role); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "Role updated successfully",
	})
}

// UpdateScope godoc
// @Summary Update a user's scope
// @Description Update permission scope of a user (admin only)
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param body body dto.UpdateScopeRequest true "New scope configuration"
// @Success 200 {object} dto.MessageResponse "Scope updated successfully"
// @Failure 400 {object} dto.ErrorResponse "Bad request"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /users/update/scope/{id} [put]
func (h *UserHandler) UpdateScope(c *gin.Context) {
	userId := c.GetString("userId")
	var req dto.UpdateScopeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	if err := h.userService.UpdateScope(userId, req.Scopes, req.IsAdded); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "Scope updated successfully",
	})
}

// Delete godoc
// @Summary Delete a user
// @Description Remove a user from the system (admin only)
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} dto.MessageResponse "User deleted successfully"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /users/delete/{id} [delete]
func (h *UserHandler) Delete(c *gin.Context) {
	userId := c.GetString("userId")

	if err := h.userService.Delete(userId); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "User deleted successfully",
	})
}
