package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vnFuhung2903/vcs-sms/dto"
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
	}
}

// SendEmail godoc
// @Summary Send container status report via email
// @Description Generates a container uptime/downtime report and sends it to the provided email address
// @Tags Report
// @Produce json
// @Param email query string true "Recipient email address"
// @Param start_time query string false "Start time in RFC3339 format (e.g. 2025-07-01T00:00:00Z)"
// @Param end_time query string false "End time in RFC3339 format (defaults to current time)"
// @Success 200 {object} dto.APIResponse "Report emailed successfully"
// @Failure 400 {object} dto.APIResponse "Invalid input or time range"
// @Failure 500 {object} dto.APIResponse "Failed to retrieve data or send email"
// @Security BearerAuth
// @Router /report/mail [get]
func (h *ReportHandler) SendEmail(c *gin.Context) {
	var req dto.ReportRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "BAD_REQUEST",
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	if req.EndTime.IsZero() {
		req.EndTime = time.Now()
	}

	if req.StartTime.After(req.EndTime) {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "BAD_REQUEST",
			Message: "Invalid request data",
			Error:   "start time cannot be after end time",
		})
		return
	}

	containers, total, err := h.containerService.View(c.Request.Context(), dto.ContainerFilter{}, 1, -1, dto.ContainerSort{
		Field: "container_id", Order: dto.Asc,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to retrieve containers",
			Error:   err.Error(),
		})
		return
	}

	var ids []string
	for _, container := range containers {
		ids = append(ids, container.ContainerId)
	}

	statusList, err := h.healthcheckService.GetEsStatus(c.Request.Context(), ids, 10000, req.StartTime, req.EndTime, dto.Asc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to retrieve healthcheck status",
			Error:   err.Error(),
		})
		return
	}

	overlapStatusList, err := h.healthcheckService.GetEsStatus(c.Request.Context(), ids, 1, req.EndTime, time.Now(), dto.Asc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to retrieve overlap healthcheck status",
			Error:   err.Error(),
		})
		return
	}

	onCount, offCount, totalUptime := h.reportService.CalculateReportStatistic(statusList, overlapStatusList, req.StartTime, req.EndTime)

	if err := h.reportService.SendEmail(c.Request.Context(), req.Email, int(total), onCount, offCount, totalUptime, req.StartTime, req.EndTime); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to send email",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "REPORT_EMAILED",
		Message: "Report emailed successfully",
	})
}
