package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vnFuhung2903/vcs-sms/dto"
	"github.com/vnFuhung2903/vcs-sms/pkg/middlewares"
	"github.com/vnFuhung2903/vcs-sms/usecases/services"
)

type HealthcheckHandler struct {
	containerService   services.IContainerService
	healthcheckService services.IHealthcheckService
}

func NewHealthcheckHandler(containerService services.IContainerService, healthcheckService services.IHealthcheckService, jwtMiddleware middlewares.IJWTMiddleware) *HealthcheckHandler {
	return &HealthcheckHandler{containerService, healthcheckService}
}

func (h *HealthcheckHandler) SetupRoutes(r *gin.Engine) {
	healthcheckRoutes := r.Group("/healthcheck")
	healthcheckRoutes.POST("/update", h.Update)
}

func (h *HealthcheckHandler) Update(c *gin.Context) {
	data, _, err := h.containerService.View(c.Request.Context(), dto.ContainerFilter{}, 1, -1, dto.ContainerSort{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	var statusList []dto.EsStatusUpdate
	for _, container := range data {
		statusList = append(statusList, dto.EsStatusUpdate{
			ContainerId: container.ContainerId,
			Status:      container.Status,
		})
	}

	if err := h.healthcheckService.UpdateStatus(c.Request.Context(), statusList); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}
	c.Status(http.StatusOK)
}
