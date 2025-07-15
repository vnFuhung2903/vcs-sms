package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/vnFuhung2903/vcs-sms/dto"
	"github.com/vnFuhung2903/vcs-sms/pkg/middlewares"
	"github.com/vnFuhung2903/vcs-sms/usecases/services"
)

type ContainerHandler struct {
	containerService services.IContainerService
	jwtMiddleware    middlewares.IJWTMiddleware
}

func NewContainerHandler(containerService services.IContainerService, jwtMiddleware middlewares.IJWTMiddleware) *ContainerHandler {
	return &ContainerHandler{containerService, jwtMiddleware}
}

func (h *ContainerHandler) SetupRoutes(r *gin.Engine) {
	containerRoutes := r.Group("/containers")
	{
		createGroup := containerRoutes.Group("", h.jwtMiddleware.RequireScope("container:create"))
		{
			createGroup.POST("/create", h.Create)
			createGroup.POST("/import", h.Import)
		}

		viewGroup := containerRoutes.Group("", h.jwtMiddleware.RequireScope("container:view"))
		{
			viewGroup.GET("/view", h.View)
			viewGroup.GET("/export", h.Export)
		}

		modifyGroup := containerRoutes.Group("", h.jwtMiddleware.RequireScope("container:update"))
		{
			modifyGroup.PUT("/update/:id", h.Update)
		}

		deleteGroup := containerRoutes.Group("", h.jwtMiddleware.RequireScope("container:delete"))
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
// @Param body body dto.CreateRequest true "Container creation request"
// @Success 200
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security ApiKeyAuth
// @Router /containers/create [post]
func (h *ContainerHandler) Create(c *gin.Context) {
	var req dto.CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	_, err := h.containerService.Create(c.Request.Context(), req.ContainerId, req.ContainerName, req.Status, req.IPv4)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}
	c.Status(http.StatusOK)
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
// @Success 200 {object} dto.ViewResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security ApiKeyAuth
// @Router /containers/view [get]
func (h *ContainerHandler) View(c *gin.Context) {
	from, err := strconv.Atoi(c.DefaultQuery("from", "1"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}
	to, err := strconv.Atoi(c.DefaultQuery("to", "10"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	var filter dto.ContainerFilter
	var sort dto.ContainerSort
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	if err := c.ShouldBindQuery(&sort); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	sort = dto.StandardizeSort(sort)
	containers, total, err := h.containerService.View(c.Request.Context(), filter, from, to, sort)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, dto.ViewResponse{
		Data:  containers,
		Total: total,
	})
}

// Update godoc
// @Summary Update a container
// @Description Update container information by ID
// @Tags containers
// @Accept json
// @Param id path string true "containerId"
// @Param body body dto.ContainerUpdate true "Update data"
// @Success 200
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security ApiKeyAuth
// @Router /containers/update/{id} [put]
func (h *ContainerHandler) Update(c *gin.Context) {
	containerId := c.Param("id")

	var updateData dto.ContainerUpdate
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	err := h.containerService.Update(c.Request.Context(), containerId, updateData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}
	c.Status(http.StatusOK)
}

// Delete godoc
// @Summary Delete a container
// @Description Delete a container by ID
// @Tags containers
// @Param id path string true "containerId"
// @Success 200
// @Failure 500 {object} dto.ErrorResponse
// @Security ApiKeyAuth
// @Router /containers/delete/{id} [delete]
func (h *ContainerHandler) Delete(c *gin.Context) {
	containerId := c.Param("id")
	err := h.containerService.Delete(c.Request.Context(), containerId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: err.Error(),
		})
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
// @Success 200 {object} dto.ImportResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security ApiKeyAuth
// @Router /containers/import [post]
func (h *ContainerHandler) Import(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}
	defer file.Close()

	result, err := h.containerService.Import(c.Request.Context(), file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, dto.ImportResponse(*result))
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
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security ApiKeyAuth
// @Router /containers/export [get]
func (h *ContainerHandler) Export(c *gin.Context) {
	from, err := strconv.Atoi(c.DefaultQuery("from", "1"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}
	to, err := strconv.Atoi(c.DefaultQuery("to", "10"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	var filter dto.ContainerFilter
	var sort dto.ContainerSort
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	if err := c.ShouldBindQuery(&sort); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	sort = dto.StandardizeSort(sort)
	data, err := h.containerService.Export(c.Request.Context(), filter, from, to, sort)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", `attachment; filename="containers.xlsx"`)
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}
