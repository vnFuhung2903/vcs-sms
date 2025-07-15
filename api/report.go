package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vnFuhung2903/vcs-sms/dto"
	"github.com/vnFuhung2903/vcs-sms/entities"
	"github.com/vnFuhung2903/vcs-sms/pkg/middlewares"
	"github.com/vnFuhung2903/vcs-sms/usecases/services"
)

type ReportHandler struct {
	containerService   services.IContainerService
	healthcheckService services.IHealthcheckService
	reportService      services.IReportService
	jwtMiddleware      middlewares.IJWTMiddleware
}

func NewReportHandler(containerService services.IContainerService, healthcheckService services.IHealthcheckService, reportService services.IReportService, jwtMiddleware middlewares.IJWTMiddleware) *ReportHandler {
	return &ReportHandler{containerService, healthcheckService, reportService, jwtMiddleware}
}

func (h *ReportHandler) SetupRoutes(r *gin.Engine) {
	reportRoutes := r.Group("/report", h.jwtMiddleware.RequireScope("report:mail"))
	{
		reportRoutes.GET("/mail", h.SendEmail)
		reportRoutes.POST("/update", h.Update)
	}
}

// SendEmail godoc
// @Summary Email the report to a user
// @Description Email the report of container management to a user (admin only)
// @Tags report
// @Produce json
// @Param id path string true "userId"
// @Success 200
// @Failure 500 {object} dto.ErrorResponse
// @Security ApiKeyAuth
// @Router /report/mail/{id} [get]
func (h *ReportHandler) SendEmail(c *gin.Context) {
	var req dto.ReportRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	filter := dto.ContainerFilter{}
	sort := dto.StandardizeSort(dto.ContainerSort{})

	data, total, err := h.containerService.View(c.Request.Context(), filter, 1, -1, sort)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	var ids []string
	onCount, offCount := 0, 0
	for _, container := range data {
		if container.Status == entities.ContainerOn {
			onCount++
		} else {
			offCount++
		}
		ids = append(ids, container.ContainerId)
	}

	results, err := h.healthcheckService.GetEsStatus(c.Request.Context(), ids, 200, req.StartTime, req.EndTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	totalUptime := 0.0
	for _, id := range ids {
		uptime := h.healthcheckService.CalculateUptimePercentage(results[id], req.StartTime, req.EndTime)
		totalUptime += uptime
	}

	response := dto.ReportResponse{
		ContainerCount:    int(total),
		ContainerOnCount:  onCount,
		ContainerOffCount: offCount,
		TotalUptime:       totalUptime,
		StartTime:         req.StartTime,
		EndTime:           req.EndTime,
	}

	if err := h.reportService.SendEmail(c.Request.Context(), &response, req.Email, req.StartTime, req.EndTime); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *ReportHandler) Update(c *gin.Context) {
	filter := dto.ContainerFilter{}
	sort := dto.StandardizeSort(dto.ContainerSort{})

	data, _, err := h.containerService.View(c.Request.Context(), filter, 1, -1, sort)
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
