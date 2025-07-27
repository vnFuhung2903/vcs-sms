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
// @Success 200 {object} dto.MessageResponse "Container created successfully"
// @Failure 400 {object} dto.APIResponse "Bad request"
// @Failure 500 {object} dto.APIResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /containers/create [post]
func (h *ContainerHandler) Create(c *gin.Context) {
	var req dto.CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "BAD_REQUEST",
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	_, err := h.containerService.Create(c.Request.Context(), req.ContainerName, req.ImageName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to create container",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, dto.APIResponse{
		Success: true,
		Code:    "CONTAINER_CREATED",
		Message: "Container created successfully",
	})
}

// View godoc
// @Summary View containers
// @Description View container list with pagination, filter, and sort
// @Tags containers
// @Produce json
// @Param from query int false "From index (default 1)" default(1)
// @Param to query int false "To index (default -1 for all)" default(-1)
// @Param status query string false "Filter by status"
// @Param field query string false "Sort by field"
// @Param order query string false "Sort order (asc or desc)" Enums(asc, desc)
// @Success 200 {object} dto.ViewResponse "Successful response with container list"
// @Failure 400 {object} dto.APIResponse "Bad request"
// @Failure 500 {object} dto.APIResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /containers/view [get]
func (h *ContainerHandler) View(c *gin.Context) {
	from, err := strconv.Atoi(c.DefaultQuery("from", "1"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "BAD_REQUEST",
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}
	to, err := strconv.Atoi(c.DefaultQuery("to", "-1"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "BAD_REQUEST",
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	var filter dto.ContainerFilter
	var sort dto.ContainerSort
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "BAD_REQUEST",
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	if err := c.ShouldBindQuery(&sort); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "BAD_REQUEST",
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	containers, total, err := h.containerService.View(c.Request.Context(), filter, from, to, sort)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to retrieve containers",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "CONTAINERS_RETRIEVED",
		Message: "Containers retrieved successfully",
		Data: dto.ViewResponse{
			Data:  containers,
			Total: total,
		},
	})
}

// Update godoc
// @Summary Update a container
// @Description Update container information by ID
// @Tags containers
// @Accept json
// @Produce json
// @Param id path string true "Container ID"
// @Param body body dto.ContainerUpdate true "Update data"
// @Success 200 {object} dto.MessageResponse "Container updated successfully"
// @Failure 400 {object} dto.APIResponse "Bad request"
// @Failure 500 {object} dto.APIResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /containers/update/{id} [put]
func (h *ContainerHandler) Update(c *gin.Context) {
	containerId := c.Param("id")

	var updateData dto.ContainerUpdate
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "BAD_REQUEST",
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	err := h.containerService.Update(c.Request.Context(), containerId, updateData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to update container",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "CONTAINER_UPDATED",
		Message: "Container updated successfully",
	})
}

// Delete godoc
// @Summary Delete a container
// @Description Delete a container by ID
// @Tags containers
// @Produce json
// @Param id path string true "Container ID"
// @Success 200 {object} dto.MessageResponse "Container deleted successfully"
// @Failure 500 {object} dto.APIResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /containers/delete/{id} [delete]
func (h *ContainerHandler) Delete(c *gin.Context) {
	containerId := c.Param("id")
	err := h.containerService.Delete(c.Request.Context(), containerId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to delete container",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "CONTAINER_DELETED",
		Message: "Container deleted successfully",
	})
}

// Import godoc
// @Summary Import containers from file
// @Description Import containers using an Excel file upload
// @Tags containers
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Excel file to import containers"
// @Success 200 {object} dto.ImportResponse "Import result with success and error counts"
// @Failure 400 {object} dto.APIResponse "Bad request"
// @Failure 500 {object} dto.APIResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /containers/import [post]
func (h *ContainerHandler) Import(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "BAD_REQUEST",
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}
	defer file.Close()

	result, err := h.containerService.Import(c.Request.Context(), file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to import containers",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "CONTAINERS_IMPORTED",
		Message: "Containers imported successfully",
		Data:    dto.ImportResponse(*result),
	})
}

// Export godoc
// @Summary Export containers
// @Description Export containers to Excel with optional filters and sort
// @Tags containers
// @Produce application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Param from query int false "From index (default 1)" default(1)
// @Param to query int false "To index (default -1 for all)" default(-1)
// @Param status query string false "Filter by status"
// @Param field query string false "Sort by field"
// @Param order query string false "Sort order (asc or desc)" Enums(asc, desc)
// @Success 200 {file} binary "Excel file containing container data"
// @Failure 400 {object} dto.APIResponse "Bad request"
// @Failure 500 {object} dto.APIResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /containers/export [get]
func (h *ContainerHandler) Export(c *gin.Context) {
	from, err := strconv.Atoi(c.DefaultQuery("from", "1"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "BAD_REQUEST",
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}
	to, err := strconv.Atoi(c.DefaultQuery("to", "-1"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "BAD_REQUEST",
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	var filter dto.ContainerFilter
	var sort dto.ContainerSort
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "BAD_REQUEST",
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	if err := c.ShouldBindQuery(&sort); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "BAD_REQUEST",
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	data, err := h.containerService.Export(c.Request.Context(), filter, from, to, sort)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to export containers",
			Error:   err.Error(),
		})
		return
	}
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", `attachment; filename="containers.xlsx"`)
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}
