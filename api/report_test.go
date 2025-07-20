package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"github.com/vnFuhung2903/vcs-sms/dto"
	"github.com/vnFuhung2903/vcs-sms/entities"
	"github.com/vnFuhung2903/vcs-sms/mocks/middlewares"
	"github.com/vnFuhung2903/vcs-sms/mocks/services"
)

type ReportHandlerSuite struct {
	suite.Suite
	ctrl                   *gomock.Controller
	mockContainerService   *services.MockIContainerService
	mockHealthcheckService *services.MockIHealthcheckService
	mockReportService      *services.MockIReportService
	mockJWTMiddleware      *middlewares.MockIJWTMiddleware
	handler                *ReportHandler
	router                 *gin.Engine
	ctx                    context.Context
}

func (s *ReportHandlerSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockContainerService = services.NewMockIContainerService(s.ctrl)
	s.mockHealthcheckService = services.NewMockIHealthcheckService(s.ctrl)
	s.mockReportService = services.NewMockIReportService(s.ctrl)
	s.mockJWTMiddleware = middlewares.NewMockIJWTMiddleware(s.ctrl)
	s.ctx = context.Background()

	s.mockJWTMiddleware.EXPECT().
		RequireScope("report:mail").
		Return(func(c *gin.Context) {
			c.Next()
		}).
		AnyTimes()

	s.handler = NewReportHandler(s.mockContainerService, s.mockHealthcheckService, s.mockReportService, s.mockJWTMiddleware)

	gin.SetMode(gin.TestMode)
	s.router = gin.New()
	s.handler.SetupRoutes(s.router)
}

func (s *ReportHandlerSuite) TearDownTest() {
	s.ctrl.Finish()
}

func TestReportHandlerSuite(t *testing.T) {
	suite.Run(t, new(ReportHandlerSuite))
}

func (s *ReportHandlerSuite) TestSendEmail() {
	startTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)

	containers := []*entities.Container{
		{ContainerId: "container1", ContainerName: "test1", Status: "running"},
		{ContainerId: "container2", ContainerName: "test2", Status: "stopped"},
	}

	esResults := map[string][]dto.EsStatus{
		"container1": {{ContainerId: "container1", Status: "running"}},
		"container2": {{ContainerId: "container2", Status: "stopped"}},
	}

	s.mockContainerService.EXPECT().
		View(gomock.Any(), dto.ContainerFilter{}, 1, -1, dto.ContainerSort{Field: "container_id", Order: "desc"}).
		Return(containers, int64(2), nil)

	s.mockHealthcheckService.EXPECT().
		GetEsStatus(gomock.Any(), []string{"container1", "container2"}, 200, startTime, endTime).
		Return(esResults, nil)

	s.mockReportService.EXPECT().
		CalculateReportStatistic(containers, esResults).
		Return(1, 1, 50.0)

	s.mockReportService.EXPECT().
		SendEmail(gomock.Any(), "test@example.com", 2, 1, 1, 50.0, startTime, endTime).
		Return(nil)

	req := httptest.NewRequest("GET", "/report/mail?email=test@example.com&start_time=2023-01-01T00:00:00Z&end_time=2023-01-02T00:00:00Z", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
}

func (s *ReportHandlerSuite) TestSendEmailInvalidQueryBinding() {
	req := httptest.NewRequest("GET", "/report/mail?start_time=invalid-datetime", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code)

	var response dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.NotEmpty(response.Error)
}

func (s *ReportHandlerSuite) TestSendEmailInvalidDateRange() {
	req := httptest.NewRequest("GET", "/report/mail?email=test@example.com&start_time=2024-01-01T00:00:00Z&end_time=2023-01-02T00:00:00Z", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code)

	var response dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.NotEmpty(response.Error)
}

func (s *ReportHandlerSuite) TestSendEmailContainerServiceError() {
	s.mockContainerService.EXPECT().
		View(gomock.Any(), dto.ContainerFilter{}, 1, -1, dto.ContainerSort{Field: "container_id", Order: "desc"}).
		Return([]*entities.Container{}, int64(0), errors.New("database connection failed"))

	req := httptest.NewRequest("GET", "/report/mail?email=test@example.com&start_time=2023-01-01T00:00:00Z&end_time=2023-01-02T00:00:00Z", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusInternalServerError, w.Code)

	var response dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal("database connection failed", response.Error)
}

func (s *ReportHandlerSuite) TestSendEmailHealthcheckServiceError() {
	startTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)

	containers := []*entities.Container{
		{ContainerId: "container1", ContainerName: "test1", Status: "running"},
	}

	s.mockContainerService.EXPECT().
		View(gomock.Any(), dto.ContainerFilter{}, 1, -1, dto.ContainerSort{Field: "container_id", Order: "desc"}).
		Return(containers, int64(1), nil)

	s.mockHealthcheckService.EXPECT().
		GetEsStatus(gomock.Any(), []string{"container1"}, 200, startTime, endTime).
		Return(map[string][]dto.EsStatus{}, errors.New("elasticsearch connection failed"))

	req := httptest.NewRequest("GET", "/report/mail?email=test@example.com&start_time=2023-01-01T00:00:00Z&end_time=2023-01-02T00:00:00Z", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusInternalServerError, w.Code)

	var response dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal("elasticsearch connection failed", response.Error)
}

func (s *ReportHandlerSuite) TestSendEmailSendEmailServiceError() {
	startTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)

	containers := []*entities.Container{
		{ContainerId: "container1", ContainerName: "test1", Status: "running"},
	}

	esResults := map[string][]dto.EsStatus{
		"container1": {{ContainerId: "container1", Status: "running"}},
	}

	s.mockContainerService.EXPECT().
		View(gomock.Any(), dto.ContainerFilter{}, 1, -1, dto.ContainerSort{Field: "container_id", Order: "desc"}).
		Return(containers, int64(1), nil)

	s.mockHealthcheckService.EXPECT().
		GetEsStatus(gomock.Any(), []string{"container1"}, 200, startTime, endTime).
		Return(esResults, nil)

	s.mockReportService.EXPECT().
		CalculateReportStatistic(containers, esResults).
		Return(1, 0, 100.0)

	s.mockReportService.EXPECT().
		SendEmail(gomock.Any(), "test@example.com", 1, 1, 0, 100.0, startTime, endTime).
		Return(errors.New("email service unavailable"))

	req := httptest.NewRequest("GET", "/report/mail?email=test@example.com&start_time=2023-01-01T00:00:00Z&end_time=2023-01-02T00:00:00Z", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusInternalServerError, w.Code)

	var response dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal("email service unavailable", response.Error)
}
