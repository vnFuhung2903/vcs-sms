package workers

import (
	"errors"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"github.com/vnFuhung2903/vcs-sms/dto"
	"github.com/vnFuhung2903/vcs-sms/entities"
	"github.com/vnFuhung2903/vcs-sms/mocks/logger"
	"github.com/vnFuhung2903/vcs-sms/mocks/middlewares"
	"github.com/vnFuhung2903/vcs-sms/mocks/services"
)

type ReportHandlerSuite struct {
	suite.Suite
	ctrl                   *gomock.Controller
	reportWorker           IReportkWorker
	mockContainerService   *services.MockIContainerService
	mockHealthcheckService *services.MockIHealthcheckService
	mockReportService      *services.MockIReportService
	mockJWTMiddleware      *middlewares.MockIJWTMiddleware
	mockLogger             *logger.MockILogger
	router                 *gin.Engine
}

func (s *ReportHandlerSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockContainerService = services.NewMockIContainerService(s.ctrl)
	s.mockHealthcheckService = services.NewMockIHealthcheckService(s.ctrl)
	s.mockReportService = services.NewMockIReportService(s.ctrl)
	s.mockJWTMiddleware = middlewares.NewMockIJWTMiddleware(s.ctrl)
	s.mockLogger = logger.NewMockILogger(s.ctrl)

	s.mockJWTMiddleware.EXPECT().
		RequireScope("report:mail").
		Return(func(c *gin.Context) {
			c.Next()
		}).
		AnyTimes()

	s.reportWorker = NewReportkWorker(s.mockContainerService, s.mockHealthcheckService, s.mockReportService, "test@example.com", s.mockLogger, 2*time.Second)
}

func (s *ReportHandlerSuite) TearDownTest() {
	s.ctrl.Finish()
}

func TestReportHandlerSuite(t *testing.T) {
	suite.Run(t, new(ReportHandlerSuite))
}

func (s *ReportHandlerSuite) TestSendEmail() {
	baseTime := time.Now()

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

	s.mockLogger.EXPECT().Info("daily report emailed successfully").AnyTimes()
	s.mockLogger.EXPECT().Info("daily report workers stopped").AnyTimes()

	s.reportWorker.Start(1)
	time.Sleep(3 * time.Second)

	s.reportWorker.Stop()
}

func (s *ReportHandlerSuite) TestSendEmailContainerServiceError() {
	s.mockContainerService.EXPECT().
		View(gomock.Any(), dto.ContainerFilter{}, 1, -1, dto.ContainerSort{Field: "container_id", Order: dto.Asc}).
		Return([]*entities.Container{}, int64(0), errors.New("container service error"))

	s.mockLogger.EXPECT().Error("failed to retrieve containers", gomock.Any()).AnyTimes()
	s.mockLogger.EXPECT().Info("daily report workers stopped").AnyTimes()

	s.reportWorker.Start(1)
	time.Sleep(3 * time.Second)

	s.reportWorker.Stop()
}

func (s *ReportHandlerSuite) TestSendEmailHealthcheckServiceError() {
	containers := []*entities.Container{
		{ContainerId: "container1", ContainerName: "test1", Status: entities.ContainerOn},
	}

	s.mockContainerService.EXPECT().
		View(gomock.Any(), dto.ContainerFilter{}, 1, -1, dto.ContainerSort{Field: "container_id", Order: dto.Asc}).
		Return(containers, int64(len(containers)), nil)

	s.mockHealthcheckService.EXPECT().
		GetEsStatus(gomock.Any(), []string{"container1"}, 10000, gomock.Any(), gomock.Any(), dto.Asc).
		Return(map[string][]dto.EsStatus{}, errors.New("elasticsearch error"))

	s.mockLogger.EXPECT().Error("failed to retrieve elasticsearch status", gomock.Any()).AnyTimes()
	s.mockLogger.EXPECT().Info("daily report workers stopped").AnyTimes()

	s.reportWorker.Start(1)
	time.Sleep(3 * time.Second)

	s.reportWorker.Stop()
}

func (s *ReportHandlerSuite) TestSendEmailHealthcheckServiceOverlapError() {
	baseTime := time.Now()
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

	s.mockLogger.EXPECT().Error("failed to retrieve elasticsearch status", gomock.Any()).AnyTimes()
	s.mockLogger.EXPECT().Info("daily report workers stopped").AnyTimes()

	s.reportWorker.Start(1)
	time.Sleep(3 * time.Second)

	s.reportWorker.Stop()
}

func (s *ReportHandlerSuite) TestSendEmailSendEmailServiceError() {
	baseTime := time.Now()
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

	s.mockLogger.EXPECT().Error("failed to email daily report", gomock.Any()).AnyTimes()
	s.mockLogger.EXPECT().Info("daily report workers stopped").AnyTimes()

	s.reportWorker.Start(1)
	time.Sleep(3 * time.Second)

	s.reportWorker.Stop()
}
