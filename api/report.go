package api

import (
	"net/http"

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

	if req.StartTime.After(req.EndTime) {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "start time must be before end time",
		})
		return
	}

	data, total, err := h.containerService.View(c.Request.Context(), dto.ContainerFilter{}, 1, -1, dto.ContainerSort{Field: "container_id", Order: "desc"})
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	var ids []string
	for _, container := range data {
		ids = append(ids, container.ContainerId)
	}

	results, err := h.healthcheckService.GetEsStatus(c.Request.Context(), ids, 200, req.StartTime, req.EndTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	onCount, offCount, totalUptime := h.reportService.CalculateReportStatistic(data, results)

	if err := h.reportService.SendEmail(c.Request.Context(), req.Email, int(total), onCount, offCount, totalUptime, req.StartTime, req.EndTime); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	c.Status(http.StatusOK)
}
