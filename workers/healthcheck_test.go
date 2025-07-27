package workers

import (
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"github.com/vnFuhung2903/vcs-sms/dto"
	"github.com/vnFuhung2903/vcs-sms/entities"
	"github.com/vnFuhung2903/vcs-sms/mocks/docker"
	"github.com/vnFuhung2903/vcs-sms/mocks/logger"
	"github.com/vnFuhung2903/vcs-sms/mocks/services"
)

type HealthcheckWorkerSuite struct {
	suite.Suite
	ctrl                   *gomock.Controller
	healthcheckWorker      IHealthcheckWorker
	mockDockerClient       *docker.MockIDockerClient
	mockContainerService   *services.MockIContainerService
	mockHealthcheckService *services.MockIHealthcheckService
	mockLogger             *logger.MockILogger
}

func (s *HealthcheckWorkerSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockDockerClient = docker.NewMockIDockerClient(s.ctrl)
	s.mockContainerService = services.NewMockIContainerService(s.ctrl)
	s.mockHealthcheckService = services.NewMockIHealthcheckService(s.ctrl)
	s.mockLogger = logger.NewMockILogger(s.ctrl)

	s.healthcheckWorker = NewHealthcheckWorker(
		s.mockDockerClient,
		s.mockContainerService,
		s.mockHealthcheckService,
		s.mockLogger,
		2*time.Second,
	)
}

func (s *HealthcheckWorkerSuite) TearDownTest() {
	s.ctrl.Finish()
}

func TestHealthcheckWorkerSuite(t *testing.T) {
	suite.Run(t, new(HealthcheckWorkerSuite))
}

func (s *HealthcheckWorkerSuite) TestHealthcheckWorkerStatusChange() {
	containers := []*entities.Container{
		{ContainerId: "1", ContainerName: "container1", Status: entities.ContainerOn},
	}

	statusList := []dto.EsStatusUpdate{
		{ContainerId: "1", Status: entities.ContainerOff},
	}

	s.mockContainerService.EXPECT().
		View(gomock.Any(), gomock.Any(), 1, -1, gomock.Any()).
		Return(containers, int64(1), nil)

	s.mockDockerClient.EXPECT().
		GetStatus(gomock.Any(), "1").
		Return(entities.ContainerOff)

	s.mockContainerService.EXPECT().
		Update(gomock.Any(), "1", dto.ContainerUpdate{Status: entities.ContainerOff}).
		Return(nil)

	s.mockHealthcheckService.EXPECT().
		UpdateStatus(gomock.Any(), statusList, gomock.Any()).
		Return(nil)

	s.mockLogger.EXPECT().Info("elasticsearch status updated successfully").AnyTimes()
	s.mockLogger.EXPECT().Info("elasticsearch status workers stopped").AnyTimes()

	s.healthcheckWorker.Start(1)
	time.Sleep(3 * time.Second)

	s.healthcheckWorker.Stop()
}

func (s *HealthcheckWorkerSuite) TestHealthcheckWorkerContainerViewError() {
	s.mockContainerService.EXPECT().
		View(gomock.Any(), gomock.Any(), 1, -1, gomock.Any()).
		Return(nil, int64(0), errors.New("view error"))

	s.mockLogger.EXPECT().Error("failed to view containers", gomock.Any()).AnyTimes()
	s.mockLogger.EXPECT().Info("elasticsearch status workers stopped").AnyTimes()

	s.healthcheckWorker.Start(1)
	time.Sleep(3 * time.Second)

	s.healthcheckWorker.Stop()
}

func (s *HealthcheckWorkerSuite) TestHealthcheckWorkerContainerUpdateError() {
	containers := []*entities.Container{
		{ContainerId: "1", ContainerName: "container1", Status: entities.ContainerOn},
	}

	s.mockContainerService.EXPECT().
		View(gomock.Any(), gomock.Any(), 1, -1, gomock.Any()).
		Return(containers, int64(1), nil)

	s.mockDockerClient.EXPECT().
		GetStatus(gomock.Any(), "1").
		Return(entities.ContainerOff)

	s.mockContainerService.EXPECT().
		Update(gomock.Any(), "1", dto.ContainerUpdate{Status: entities.ContainerOff}).
		Return(errors.New("update error"))

	s.mockHealthcheckService.EXPECT().
		UpdateStatus(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil)

	s.mockLogger.EXPECT().Error("failed to update container", gomock.Any()).AnyTimes()
	s.mockLogger.EXPECT().Info("elasticsearch status updated successfully").AnyTimes()
	s.mockLogger.EXPECT().Info("elasticsearch status workers stopped").AnyTimes()

	s.healthcheckWorker.Start(1)
	time.Sleep(3 * time.Second)

	s.healthcheckWorker.Stop()
}

func (s *HealthcheckWorkerSuite) TestHealthcheckWorkerNoStatusChange() {
	containers := []*entities.Container{
		{ContainerId: "1", ContainerName: "container1", Status: entities.ContainerOn},
	}

	s.mockContainerService.EXPECT().
		View(gomock.Any(), gomock.Any(), 1, -1, gomock.Any()).
		Return(containers, int64(1), nil)

	s.mockDockerClient.EXPECT().
		GetStatus(gomock.Any(), "1").
		Return(entities.ContainerOn)

	s.mockHealthcheckService.EXPECT().
		UpdateStatus(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil)

	s.mockLogger.EXPECT().Info("elasticsearch status updated successfully").AnyTimes()
	s.mockLogger.EXPECT().Info("elasticsearch status workers stopped").AnyTimes()

	s.healthcheckWorker.Start(1)
	time.Sleep(3 * time.Second)

	s.healthcheckWorker.Stop()
}

func (s *HealthcheckWorkerSuite) TestHealthcheckWorkerUpdateStatusError() {
	containers := []*entities.Container{
		{ContainerId: "1", ContainerName: "container1", Status: entities.ContainerOn},
	}

	s.mockContainerService.EXPECT().
		View(gomock.Any(), gomock.Any(), 1, -1, gomock.Any()).
		Return(containers, int64(1), nil)

	s.mockDockerClient.EXPECT().
		GetStatus(gomock.Any(), "1").
		Return(entities.ContainerOn)

	s.mockHealthcheckService.EXPECT().
		UpdateStatus(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(errors.New("update status error"))

	s.mockLogger.EXPECT().Error("failed to update elasticsearch status", gomock.Any()).AnyTimes()
	s.mockLogger.EXPECT().Info("elasticsearch status workers stopped").AnyTimes()

	s.healthcheckWorker.Start(1)
	time.Sleep(3 * time.Second)

	s.healthcheckWorker.Stop()
}
