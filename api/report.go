package api

import (
	"errors"
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
// @Summary Send email report to user
// @Description Send container management report via email to specified user
// @Tags report
// @Produce json
// @Param email query string true "Email address to send report to"
// @Param start_time query string false "Start time for report"
// @Param end_time query string false "End time for report (defaults to now)"
// @Success 200 {object} dto.MessageResponse "Email sent successfully"
// @Failure 400 {object} dto.APIResponse "Bad request"
// @Failure 500 {object} dto.APIResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /report/mail [get]
func (h *ReportHandler) SendEmail(c *gin.Context) {
	var req dto.ReportRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "BAD_REQUEST",
			Message: "Invalid request data",
			Error:   err,
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
			Error:   errors.New("start time cannot be after end time"),
		})
		return
	}

	data, total, err := h.containerService.View(c.Request.Context(), dto.ContainerFilter{}, 1, -1, dto.ContainerSort{Field: "container_id", Order: "desc"})
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to retrieve containers",
			Error:   err,
		})
		return
	}

	var ids []string
	for _, container := range data {
		ids = append(ids, container.ContainerId)
	}

	results, err := h.healthcheckService.GetEsStatus(c.Request.Context(), ids, 200, req.StartTime, req.EndTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to retrieve healthcheck status",
			Error:   err,
		})
		return
	}

	onCount, offCount, totalUptime := h.reportService.CalculateReportStatistic(data, results)

	if err := h.reportService.SendEmail(c.Request.Context(), req.Email, int(total), onCount, offCount, totalUptime, req.StartTime, req.EndTime); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to send email",
			Error:   err,
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "EMAIL_SENT",
		Message: "Email sent successfully",
	})
}
