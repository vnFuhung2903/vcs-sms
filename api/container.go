package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vnFuhung2903/vcs-sms/entities"
	"github.com/vnFuhung2903/vcs-sms/usecases/services"
	"github.com/vnFuhung2903/vcs-sms/utils/middlewares"
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
		createGroup := containerRoutes.Group("", middlewares.RequireScope("container:create"))
		{
			createGroup.POST("/create", h.Create)
			createGroup.POST("/import", h.Import)
		}

		viewGroup := containerRoutes.Group("", middlewares.RequireScope("container:view"))
		{
			viewGroup.GET("/view", h.View)
			viewGroup.GET("/export", h.Export)
		}

		modifyGroup := containerRoutes.Group("", middlewares.RequireScope("container:update"))
		{
			modifyGroup.PUT("/update/:id", h.Update)
		}

		deleteGroup := containerRoutes.Group("", middlewares.RequireScope("container:delete"))
		{
			deleteGroup.DELETE("/delete/:id", h.Delete)
		}
	}
}

// Create godoc
// @Summary Create new container
// @Description Create a container with ID, name, status, and IPv4
// @Tags containers
// @Accept json
// @Produce json
// @Param body body CreateRequest true "Container creation request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} entities.ErrorResponse
// @Failure 500 {object} entities.ErrorResponse
// @Security ApiKeyAuth
// @Router /containers/create [post]
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

// View godoc
// @Summary View containers
// @Description View container list with pagination, filter, and sort
// @Tags containers
// @Produce json
// @Param from query int false "From index (default 1)"
// @Param to query int false "To index (default 10)"
// @Param status query string false "Filter by status"
// @Param sort_by query string false "Sort by field"
// @Param sort_order query string false "Sort order (asc or desc)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} entities.ErrorResponse
// @Failure 500 {object} entities.ErrorResponse
// @Security ApiKeyAuth
// @Router /containers/view [get]
func (h *ContainerHandler) View(c *gin.Context) {
	from, _ := strconv.Atoi(c.DefaultQuery("from", "1"))
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

	sort = entities.StandardizeSort(sort)
	containers, total, err := h.containerService.View(c.Request.Context(), filter, from, to, sort)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": containers, "total": total})
}

// Update godoc
// @Summary Update a container
// @Description Update container information by ID
// @Tags containers
// @Accept json
// @Param id path string true "Container ID"
// @Param body body entities.ContainerUpdate true "Update data"
// @Success 200
// @Failure 400 {object} entities.ErrorResponse
// @Failure 500 {object} entities.ErrorResponse
// @Security ApiKeyAuth
// @Router /containers/update/{id} [put]
func (h *ContainerHandler) Update(c *gin.Context) {
	containerID := c.Param("id")

	var updateData entities.ContainerUpdate
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

// Delete godoc
// @Summary Delete a container
// @Description Delete a container by ID
// @Tags containers
// @Param id path string true "Container ID"
// @Success 200
// @Failure 500 {object} entities.ErrorResponse
// @Security ApiKeyAuth
// @Router /containers/delete/{id} [delete]
func (h *ContainerHandler) Delete(c *gin.Context) {
	containerID := c.Param("id")
	err := h.containerService.Delete(c.Request.Context(), containerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

// Import godoc
// @Summary Import containers from file
// @Description Import containers using an Excel file upload
// @Tags containers
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Excel file"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} entities.ErrorResponse
// @Failure 500 {object} entities.ErrorResponse
// @Security ApiKeyAuth
// @Router /containers/import [post]
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

// Export godoc
// @Summary Export containers
// @Description Export containers to Excel with optional filters and sort
// @Tags containers
// @Produce application/octet-stream
// @Param from query int false "From index (default 1)"
// @Param to query int false "To index (default 10)"
// @Param status query string false "Filter by status"
// @Param sort_by query string false "Sort by field"
// @Param sort_order query string false "Sort order (asc or desc)"
// @Success 200 {file} file
// @Failure 400 {object} entities.ErrorResponse
// @Failure 500 {object} entities.ErrorResponse
// @Security ApiKeyAuth
// @Router /containers/export [get]
func (h *ContainerHandler) Export(c *gin.Context) {
	from, _ := strconv.Atoi(c.DefaultQuery("from", "1"))
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

	sort = entities.StandardizeSort(sort)
	data, err := h.containerService.Export(c.Request.Context(), filter, from, to, sort)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", `attachment; filename="containers.xlsx"`)
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}
