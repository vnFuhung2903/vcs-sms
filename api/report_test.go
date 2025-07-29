package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
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
}

func (s *ReportHandlerSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockContainerService = services.NewMockIContainerService(s.ctrl)
	s.mockHealthcheckService = services.NewMockIHealthcheckService(s.ctrl)
	s.mockReportService = services.NewMockIReportService(s.ctrl)
	s.mockJWTMiddleware = middlewares.NewMockIJWTMiddleware(s.ctrl)

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
	baseTime := time.Now()
	endTime := baseTime
	startTime := baseTime.Add(-4 * time.Hour)

	statusList := map[string][]dto.EsStatus{
		"container1": {
			{ContainerId: "container1", Status: entities.ContainerOn, Uptime: int64(3600), LastUpdated: baseTime.Add(-210 * time.Minute)},
			{ContainerId: "container1", Status: entities.ContainerOff, Uptime: int64(1800), LastUpdated: baseTime.Add(-3 * time.Hour)},
			{ContainerId: "container1", Status: entities.ContainerOn, Uptime: int64(3600), LastUpdated: baseTime.Add(-2 * time.Hour)},
		},
		"container2": {
			{ContainerId: "container2", Status: entities.ContainerOff, Uptime: int64(7200), LastUpdated: baseTime.Add(-1 * time.Minute)},
		},
	}

	overlapStatusList := map[string][]dto.EsStatus{
		"container1": {},
		"container2": {},
	}

	containers := []*entities.Container{
		{ContainerId: "container1", ContainerName: "test1", Status: entities.ContainerOn},
		{ContainerId: "container2", ContainerName: "test2", Status: entities.ContainerOff},
	}

	s.mockContainerService.EXPECT().
		View(gomock.Any(), dto.ContainerFilter{}, 1, -1, dto.ContainerSort{Field: "container_id", Order: dto.Asc}).
		Return(containers, int64(len(containers)), nil)

	s.mockHealthcheckService.EXPECT().
		GetEsStatus(gomock.Any(), []string{"container1", "container2"}, 10000, gomock.Any(), gomock.Any(), dto.Asc).
		Return(statusList, nil)

	s.mockHealthcheckService.EXPECT().
		GetEsStatus(gomock.Any(), []string{"container1", "container2"}, 1, gomock.Any(), gomock.Any(), dto.Asc).
		Return(overlapStatusList, nil)

	s.mockReportService.EXPECT().
		CalculateReportStatistic(statusList, overlapStatusList, gomock.Any(), gomock.Any()).
		Return(1, 1, 50.0)

	s.mockReportService.EXPECT().
		SendEmail(gomock.Any(), "test@example.com", 2, 1, 1, 50.0, gomock.Any(), gomock.Any()).
		Return(nil)

	params := url.Values{}
	params.Set("email", "test@example.com")
	params.Set("start_time", startTime.UTC().Format("2006-01-02"))
	params.Set("end_time", endTime.UTC().Format("2006-01-02"))

	req := httptest.NewRequest("GET", "/report/mail?"+params.Encode(), nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusOK, w.Code)

	var response dto.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.True(response.Success)
	s.Equal("REPORT_EMAILED", response.Code)
}

func (s *ReportHandlerSuite) TestSendEmailInvalidQueryBinding() {
	req := httptest.NewRequest("GET", "/report/mail?start_time=invalid-datetime", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusBadRequest, w.Code)

	var response dto.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.NotEmpty(response.Error)

	req = httptest.NewRequest("GET", "/report/mail?start_time=2006-01-02&end_time=invalid-datetime", nil)
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusBadRequest, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.NotEmpty(response.Error)
}

func (s *ReportHandlerSuite) TestSendEmailInvalidDateRange() {
	req := httptest.NewRequest("GET", "/report/mail?email=test@example.com&start_time=2024-01-01T00:00:00Z&end_time=2023-01-02T00:00:00Z", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusBadRequest, w.Code)

	var response dto.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.NotEmpty(response.Error)
}

func (s *ReportHandlerSuite) TestSendEmailContainerServiceError() {
	s.mockContainerService.EXPECT().
		View(gomock.Any(), dto.ContainerFilter{}, 1, -1, dto.ContainerSort{Field: "container_id", Order: dto.Asc}).
		Return([]*entities.Container{}, int64(0), errors.New("container service error"))

	req := httptest.NewRequest("GET", "/report/mail?email=test@example.com&start_time=2023-01-01T00:00:00Z", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusInternalServerError, w.Code)

	var response dto.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal("container service error", response.Error)
}

func (s *ReportHandlerSuite) TestSendEmailHealthcheckServiceError() {
	baseTime := time.Now()
	endTime := baseTime
	startTime := baseTime.Add(-4 * time.Hour)

	containers := []*entities.Container{
		{ContainerId: "container1", ContainerName: "test1", Status: entities.ContainerOn},
	}

	s.mockContainerService.EXPECT().
		View(gomock.Any(), dto.ContainerFilter{}, 1, -1, dto.ContainerSort{Field: "container_id", Order: dto.Asc}).
		Return(containers, int64(len(containers)), nil)

	s.mockHealthcheckService.EXPECT().
		GetEsStatus(gomock.Any(), []string{"container1"}, 10000, gomock.Any(), gomock.Any(), dto.Asc).
		Return(map[string][]dto.EsStatus{}, errors.New("elasticsearch error"))

	params := url.Values{}
	params.Set("email", "test@example.com")
	params.Set("start_time", startTime.UTC().Format("2006-01-02"))
	params.Set("end_time", endTime.UTC().Format("2006-01-02"))

	req := httptest.NewRequest("GET", "/report/mail?"+params.Encode(), nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusInternalServerError, w.Code)

	var response dto.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal("elasticsearch error", response.Error)
}

func (s *ReportHandlerSuite) TestSendEmailHealthcheckServiceOverlapError() {
	baseTime := time.Now()
	endTime := baseTime
	startTime := endTime.Add(-4 * time.Hour)
	statusList := map[string][]dto.EsStatus{
		"container1": {
			{ContainerId: "container1", Status: entities.ContainerOn, Uptime: int64(3600), LastUpdated: baseTime.Add(-210 * time.Minute)},
			{ContainerId: "container1", Status: entities.ContainerOff, Uptime: int64(1800), LastUpdated: baseTime.Add(-3 * time.Hour)},
			{ContainerId: "container1", Status: entities.ContainerOn, Uptime: int64(3600), LastUpdated: baseTime.Add(-2 * time.Hour)},
		},
	}

	containers := []*entities.Container{
		{ContainerId: "container1", ContainerName: "test1", Status: entities.ContainerOn},
	}

	s.mockContainerService.EXPECT().
		View(gomock.Any(), dto.ContainerFilter{}, 1, -1, dto.ContainerSort{Field: "container_id", Order: dto.Asc}).
		Return(containers, int64(len(containers)), nil)

	s.mockHealthcheckService.EXPECT().
		GetEsStatus(gomock.Any(), []string{"container1"}, 10000, gomock.Any(), gomock.Any(), dto.Asc).
		Return(statusList, nil)

	s.mockHealthcheckService.EXPECT().
		GetEsStatus(gomock.Any(), []string{"container1"}, 1, gomock.Any(), gomock.Any(), dto.Asc).
		Return(map[string][]dto.EsStatus{}, errors.New("elasticsearch error"))

	params := url.Values{}
	params.Set("email", "test@example.com")
	params.Set("start_time", startTime.UTC().Format("2006-01-02"))
	params.Set("end_time", endTime.UTC().Format("2006-01-02"))

	req := httptest.NewRequest("GET", "/report/mail?"+params.Encode(), nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusInternalServerError, w.Code)

	var response dto.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal("elasticsearch error", response.Error)
}

func (s *ReportHandlerSuite) TestSendEmailSendEmailServiceError() {
	baseTime := time.Now()
	endTime := baseTime
	startTime := endTime.Add(-4 * time.Hour)
	statusList := map[string][]dto.EsStatus{
		"container1": {
			{ContainerId: "container1", Status: entities.ContainerOn, Uptime: int64(3600), LastUpdated: baseTime.Add(-210 * time.Minute)},
			{ContainerId: "container1", Status: entities.ContainerOff, Uptime: int64(1800), LastUpdated: baseTime.Add(-3 * time.Hour)},
			{ContainerId: "container1", Status: entities.ContainerOn, Uptime: int64(3600), LastUpdated: baseTime.Add(-2 * time.Hour)},
		},
	}

	overlapStatusList := map[string][]dto.EsStatus{
		"container1": {},
	}

	containers := []*entities.Container{
		{ContainerId: "container1", ContainerName: "test1", Status: entities.ContainerOn},
	}

	s.mockContainerService.EXPECT().
		View(gomock.Any(), dto.ContainerFilter{}, 1, -1, dto.ContainerSort{Field: "container_id", Order: dto.Asc}).
		Return(containers, int64(len(containers)), nil)

	s.mockHealthcheckService.EXPECT().
		GetEsStatus(gomock.Any(), []string{"container1"}, 10000, gomock.Any(), gomock.Any(), dto.Asc).
		Return(statusList, nil)

	s.mockHealthcheckService.EXPECT().
		GetEsStatus(gomock.Any(), []string{"container1"}, 1, gomock.Any(), gomock.Any(), dto.Asc).
		Return(overlapStatusList, nil)

	s.mockReportService.EXPECT().
		CalculateReportStatistic(statusList, overlapStatusList, gomock.Any(), gomock.Any()).
		Return(1, 0, 100.0)

	s.mockReportService.EXPECT().
		SendEmail(gomock.Any(), "test@example.com", 1, 1, 0, 100.0, gomock.Any(), gomock.Any()).
		Return(errors.New("service error"))

	params := url.Values{}
	params.Set("email", "test@example.com")
	params.Set("start_time", startTime.UTC().Format("2006-01-02"))
	params.Set("end_time", endTime.UTC().Format("2006-01-02"))

	req := httptest.NewRequest("GET", "/report/mail?"+params.Encode(), nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusInternalServerError, w.Code)

	var response dto.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal("service error", response.Error)
}
