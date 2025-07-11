package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vnFuhung2903/vcs-sms/dto"
	"github.com/vnFuhung2903/vcs-sms/entities"
	"github.com/vnFuhung2903/vcs-sms/usecases/services"
	"github.com/vnFuhung2903/vcs-sms/utils/hashmap"
	"github.com/vnFuhung2903/vcs-sms/utils/middlewares"
)

type UserHandler struct {
	authService services.IAuthService
	userService services.IUserService
}

func NewUserHandler(authService services.IAuthService, userService services.IUserService) *UserHandler {
	return &UserHandler{authService, userService}
}

func (h *UserHandler) SetupRoutes(r *gin.Engine) {
	userRoutes := r.Group("/users")
	{
		userRoutes.POST("/register", h.Register)
		userRoutes.POST("/login", h.Login)

		userGroup := userRoutes.Group("", middlewares.RequireScope("user:modify"))
		{
			userGroup.PUT("/update/password/:id", h.UpdatePassword)
		}

		adminGroup := userRoutes.Group("", middlewares.RequireScope("user:manage"))
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
// @Param body body struct{Username string json:"username"; Password string json:"password"; Email string json:"email"; Role string json:"role"} true "User registration"
// @Success 200
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /users/register [post]
func (h *UserHandler) Register(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
		Role     string `json:"role"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	scopes := hashmap.UserRoleToDefaultScopes(entities.UserRole(req.Role), nil)
	if err := h.userService.Register(req.Username, req.Password, req.Email, entities.UserRole(req.Role), hashmap.ScopesToHashMap(scopes)); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	if err := h.authService.Setup(req.Username, req.Username, hashmap.ScopesToHashMap(scopes)); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}
	c.Status(http.StatusOK)
}

// Login godoc
// @Summary Login with username and password
// @Description Login and receive JWT token
// @Tags users
// @Accept json
// @Produce json
// @Param body body struct{Username string json:"username"; Password string json:"password"} true "User login"
// @Success 200
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /users/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

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

	if err := h.authService.Setup(req.Username, req.Username, user.Scopes); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}
	c.Status(http.StatusOK)
}

// UpdatePassword godoc
// @Summary Update own password
// @Description Update password of currently logged-in user
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "userId"
// @Param body body struct{Password string json:"password"} true "New password"
// @Success 200
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security ApiKeyAuth
// @Router /users/update/password/{id} [put]
func (h *UserHandler) UpdatePassword(c *gin.Context) {
	userId := c.GetString("userId")
	var req struct {
		Password string `json:"password"`
	}

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

	c.Status(http.StatusOK)
}

// UpdateRole godoc
// @Summary Update a user's role
// @Description Update role of a user (admin only)
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "userId"
// @Param body body struct{Role string json:"role"} true "New role"
// @Success 200
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security ApiKeyAuth
// @Router /users/update/role/{id} [put]
func (h *UserHandler) UpdateRole(c *gin.Context) {
	userId := c.GetString("userId")
	var req struct {
		Role string `json:"role"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	if err := h.userService.UpdateRole(userId, entities.UserRole(req.Role)); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	c.Status(http.StatusOK)
}

// UpdateScope godoc
// @Summary Update a user's scope
// @Description Update permission scope of a user (admin only)
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "userId"
// @Param body body struct{Scopes int64 json:"scope"} true "New scope bitmap"
// @Success 200
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security ApiKeyAuth
// @Router /users/update/scope/{id} [put]
func (h *UserHandler) UpdateScope(c *gin.Context) {
	userId := c.GetString("userId")
	var req struct {
		IsAdded bool     `json:"isAdded"`
		Scopes  []string `json:"scope"`
	}

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

	c.Status(http.StatusOK)
}

// Delete godoc
// @Summary Delete a user
// @Description Remove a user from the system (admin only)
// @Tags users
// @Produce json
// @Param id path string true "userId"
// @Success 200
// @Failure 500 {object} dto.ErrorResponse
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

	c.Status(http.StatusOK)
}
