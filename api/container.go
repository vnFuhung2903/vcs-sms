package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vnFuhung2903/vcs-sms/entities"
	"github.com/vnFuhung2903/vcs-sms/usecases/services"
)

type ContainerHandler struct {
	containerService services.IContainerService
}

func NewContainerHandler(containerService services.IContainerService) *ContainerHandler {
	return &ContainerHandler{containerService}
}

type CreateRequest struct {
	ContainerID   string                   `json:"container_id"`
	ContainerName string                   `json:"container_name"`
	Status        entities.ContainerStatus `json:"status"`
	IPv4          string                   `json:"ipv4"`
}

func (h *ContainerHandler) SetupRoutes(r *gin.Engine) {
	containerRoutes := r.Group("/containers")
	{
		containerRoutes.POST("/create", h.Create)
		containerRoutes.GET("", h.View)
		containerRoutes.PUT("/:id", h.Update)
		containerRoutes.DELETE("/:id", h.Delete)
		containerRoutes.POST("/import", h.Import)
		containerRoutes.GET("/export", h.Export)
	}
}

func (h *ContainerHandler) Create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	container, err := h.containerService.Create(c.Request.Context(), req.ContainerID, req.ContainerName, req.Status, req.IPv4)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"created_at": container.CreatedAt.Format(time.RFC3339),
	})

}

func (h *ContainerHandler) View(c *gin.Context) {
	from, _ := strconv.Atoi(c.DefaultQuery("from", "0"))
	to, _ := strconv.Atoi(c.DefaultQuery("to", "10"))

	var filter entities.ContainerFilter
	var sort entities.ContainerSort
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid filter params: " + err.Error()})
		return
	}

	if err := c.ShouldBindQuery(&sort); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sort params: " + err.Error()})
		return
	}

	if err := entities.ValidateSort(sort); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sort params: " + err.Error()})
		return
	}

	containers, total, err := h.containerService.View(c.Request.Context(), filter, from, to, sort)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": containers, "total": total})
}

func (h *ContainerHandler) Update(c *gin.Context) {
	containerID := c.Param("id")

	var updateData map[string]any
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.containerService.Update(c.Request.Context(), containerID, updateData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (h *ContainerHandler) Delete(c *gin.Context) {
	containerID := c.Param("id")
	err := h.containerService.Delete(c.Request.Context(), containerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (h *ContainerHandler) Import(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File upload failed: " + err.Error()})
		return
	}
	defer file.Close()

	result, err := h.containerService.Import(c.Request.Context(), file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *ContainerHandler) Export(c *gin.Context) {
	from, _ := strconv.Atoi(c.DefaultQuery("from", "0"))
	to, _ := strconv.Atoi(c.DefaultQuery("to", "10"))

	var filter entities.ContainerFilter
	var sort entities.ContainerSort
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid filter params: " + err.Error()})
		return
	}

	if err := c.ShouldBindQuery(&sort); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sort params: " + err.Error()})
		return
	}

	if err := entities.ValidateSort(sort); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sort params: " + err.Error()})
		return
	}

	data, err := h.containerService.Export(c.Request.Context(), filter, from, to, sort)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Data(http.StatusOK, "application/octet-stream", data)
}
