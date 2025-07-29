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
// @Param start_time query string true "Start date (e.g. 2006-01-02)"
// @Param end_time query string false "End date (defaults to current time)"
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

	startTime, err := time.Parse(time.RFC3339, req.StartTime+"T00:00:00Z")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "BAD_REQUEST",
			Message: "Invalid start time format",
			Error:   err.Error(),
		})
		return
	}

	var endTime time.Time
	if req.EndTime == "" {
		endTime = time.Now()
	} else {
		endTime, err = time.Parse(time.RFC3339, req.EndTime+"T23:59:59Z")
		if err != nil {
			c.JSON(http.StatusBadRequest, dto.APIResponse{
				Success: false,
				Code:    "BAD_REQUEST",
				Message: "Invalid start time format",
				Error:   err.Error(),
			})
			return
		}
	}

	if startTime.After(endTime) {
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

	statusList, err := h.healthcheckService.GetEsStatus(c.Request.Context(), ids, 10000, startTime, endTime, dto.Asc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to retrieve healthcheck status",
			Error:   err.Error(),
		})
		return
	}

	overlapStatusList, err := h.healthcheckService.GetEsStatus(c.Request.Context(), ids, 1, endTime, time.Now(), dto.Asc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to retrieve overlap healthcheck status",
			Error:   err.Error(),
		})
		return
	}

	onCount, offCount, totalUptime := h.reportService.CalculateReportStatistic(statusList, overlapStatusList, startTime, endTime)

	if err := h.reportService.SendEmail(c.Request.Context(), req.Email, int(total), onCount, offCount, totalUptime, startTime, endTime); err != nil {
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
