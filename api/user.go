package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vnFuhung2903/vcs-sms/entities"
	"github.com/vnFuhung2903/vcs-sms/usecases/services"
	"github.com/vnFuhung2903/vcs-sms/utils/middlewares"
)

type UserHandler struct {
	userService services.IUserService
}

func NewUserHandler(userService services.IUserService) *UserHandler {
	return &UserHandler{userService}
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

func (h *UserHandler) Register(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
		Role     string `json:"role"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	scopes := middlewares.UserRoleToDefaultScopes(entities.UserRole(req.Role))
	user, err := h.userService.Register(req.Username, req.Password, req.Email, entities.UserRole(req.Role), middlewares.ScopeHashMap(scopes))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	token, err := middlewares.GenerateJWT(user.ID, user.Username, scopes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token generation failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user, "token": token})
}

func (h *UserHandler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	user, err := h.userService.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	token, err := middlewares.GenerateJWT(user.ID, user.Username, []string{"user", "container:read"})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token generation failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user, "token": token})
}

func (h *UserHandler) UpdatePassword(c *gin.Context) {
	userId := c.GetString("userId")
	var req struct {
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.userService.UpdatePassword(userId, req.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password updated"})
}

func (h *UserHandler) UpdateRole(c *gin.Context) {
	userId := c.GetString("userId")
	var req struct {
		Role string `json:"role"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.userService.UpdateRole(userId, entities.UserRole(req.Role)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role updated"})
}

func (h *UserHandler) UpdateScope(c *gin.Context) {
	userId := c.GetString("userId")
	var req struct {
		Scopes int64 `json:"scope"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.userService.UpdateScope(userId, req.Scopes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Scope updated"})
}

func (h *UserHandler) Delete(c *gin.Context) {
	userId := c.GetString("userId")

	if err := h.userService.Delete(userId); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted"})
}
