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
// @Summary Send email report to user
// @Description Send container management report via email to specified user
// @Tags report
// @Produce json
// @Param email query string true "Email address to send report to"
// @Param start_time query string false "Start time for report"
// @Param end_time query string false "End time for report (defaults to now)"
// @Success 200 {object} dto.MessageResponse "Email sent successfully"
// @Failure 400 {object} dto.ErrorResponse "Bad request"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /report/mail [get]
func (h *ReportHandler) SendEmail(c *gin.Context) {
	var req dto.ReportRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	if req.EndTime.IsZero() {
		req.EndTime = time.Now()
	}

	if req.StartTime.After(req.EndTime) {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "start time must be before end time",
		})
		return
	}

	containers, total, err := h.containerService.View(c.Request.Context(), dto.ContainerFilter{}, 1, -1, dto.ContainerSort{
		Field: "created_at", Order: dto.Asc,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	var ids []string
	for _, container := range containers {
		ids = append(ids, container.ContainerId)
	}

	statusList, err := h.healthcheckService.GetEsStatus(c.Request.Context(), ids, 10000, req.StartTime, req.EndTime, dto.Dsc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	overlapStatusList, err := h.healthcheckService.GetEsStatus(c.Request.Context(), ids, 1, req.EndTime, time.Now(), dto.Asc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	onCount, offCount, totalUptime := h.reportService.CalculateReportStatistic(containers, statusList, overlapStatusList)

	if err := h.reportService.SendEmail(c.Request.Context(), req.Email, int(total), onCount, offCount, totalUptime, req.StartTime, req.EndTime); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "Email sent successfully",
	})
}
