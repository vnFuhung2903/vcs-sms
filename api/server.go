package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vnFuhung2903/vcs-sms/entities"
	"github.com/vnFuhung2903/vcs-sms/usecases/services"
)

type ServerHandler struct {
	serverService services.IServerService
}

func NewServerHandler(serverService services.IServerService) *ServerHandler {
	return &ServerHandler{serverService}
}

type CreateRequest struct {
	ServerID   string                `json:"server_id"`
	ServerName string                `json:"server_name"`
	Status     entities.ServerStatus `json:"status"`
	IPv4       string                `json:"ipv4"`
}

func (h *ServerHandler) SetupRoutes(r *gin.Engine) {
	serverRoutes := r.Group("/servers")
	{
		serverRoutes.POST("/create", h.Create)
		serverRoutes.GET("", h.View)
		serverRoutes.PUT("/:id", h.Update)
		serverRoutes.DELETE("/:id", h.Delete)
		serverRoutes.POST("/import", h.Import)
		serverRoutes.GET("/export", h.Export)
	}
}

func (h *ServerHandler) Create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	server, err := h.serverService.Create(c.Request.Context(), req.ServerID, req.ServerName, req.Status, req.IPv4)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"created_at": server.CreatedAt.Format(time.RFC3339),
	})

}

func (h *ServerHandler) View(c *gin.Context) {
	from, _ := strconv.Atoi(c.DefaultQuery("from", "0"))
	to, _ := strconv.Atoi(c.DefaultQuery("to", "10"))

	var filter entities.ServerFilter
	var sort entities.ServerSort
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid filter params: " + err.Error()})
		return
	}

	if err := c.ShouldBindQuery(&sort); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sort params: " + err.Error()})
		return
	}

	servers, total, err := h.serverService.View(c.Request.Context(), filter, from, to, sort)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": servers, "total": total})
}

func (h *ServerHandler) Update(c *gin.Context) {
	serverID := c.Param("id")

	var updateData map[string]any
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.serverService.Update(c.Request.Context(), serverID, updateData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (h *ServerHandler) Delete(c *gin.Context) {
	serverID := c.Param("id")
	err := h.serverService.Delete(c.Request.Context(), serverID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (h *ServerHandler) Import(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File upload failed: " + err.Error()})
		return
	}
	defer file.Close()

	result, err := h.serverService.Import(c.Request.Context(), file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *ServerHandler) Export(c *gin.Context) {
	from, _ := strconv.Atoi(c.DefaultQuery("from", "0"))
	to, _ := strconv.Atoi(c.DefaultQuery("to", "10"))

	var filter entities.ServerFilter
	var sort entities.ServerSort
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid filter params: " + err.Error()})
		return
	}

	if err := c.ShouldBindQuery(&sort); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sort params: " + err.Error()})
		return
	}

	data, err := h.serverService.Export(c.Request.Context(), filter, from, to, sort)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Data(http.StatusOK, "application/octet-stream", data)
}
