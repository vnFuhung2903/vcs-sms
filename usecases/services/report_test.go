package services

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"github.com/vnFuhung2903/vcs-sms/dto"
	"github.com/vnFuhung2903/vcs-sms/entities"
	"github.com/vnFuhung2903/vcs-sms/mocks/logger"
	"github.com/vnFuhung2903/vcs-sms/pkg/env"
)

type ReportServiceTestSuite struct {
	suite.Suite
	ctrl          *gomock.Controller
	reportService IReportService
	logger        *logger.MockILogger
	ctx           context.Context
	sampleReport  *dto.ReportResponse
}

func (s *ReportServiceTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.logger = logger.NewMockILogger(s.ctrl)

	s.reportService = NewReportService(s.logger, env.GomailEnv{
		MailUsername: "test@gmail.com",
		MailPassword: "testpass",
	})
	s.ctx = context.Background()

	s.sampleReport = &dto.ReportResponse{
		ContainerCount:    10,
		ContainerOnCount:  7,
		ContainerOffCount: 3,
		TotalUptime:       24.5,
		StartTime:         time.Now().Add(-24 * time.Hour),
		EndTime:           time.Now(),
	}

	// Create test HTML template file
	err := os.MkdirAll("html", 0755)
	if err != nil {
		s.T().Fatal("Failed to create html directory:", err)
	}

	htmlContent := `<!DOCTYPE html>
<html>
<head><title>Test Report</title></head>
<body>
	<h1>Daily Container Report</h1>
	<p>{{ .StartTime | formatTime }} - {{ .EndTime | formatTime }}</p>
	<p>Total Container: {{ .ContainerCount }}</p>
	<p>Online Containers: {{ .ContainerOnCount }}</p>
	<p>Offline Containers: {{ .ContainerOffCount }}</p>
	<p>Total Uptime: {{ .TotalUptime }}h</p>
</body>
</html>`

	err = os.WriteFile("html/email.html", []byte(htmlContent), 0644)
	if err != nil {
		s.T().Fatal("Failed to create html file:", err)
	}
}

func (s *ReportServiceTestSuite) TearDownTest() {
	os.RemoveAll("html")
	s.ctrl.Finish()
}

func TestReportService(t *testing.T) {
	suite.Run(t, new(ReportServiceTestSuite))
}

func (s *ReportServiceTestSuite) TestSendEmail() {
	s.reportService = NewReportService(s.logger, env.GomailEnv{
		MailUsername: "hung29032004@gmail.com",
		MailPassword: "ltqisrdmlbbnhwzn",
	})
	s.logger.EXPECT().Info("Report sent successfully", gomock.Any(), gomock.Any()).Times(1)
	err := s.reportService.SendEmail(s.ctx, "hung29032004@gmail.com", 10, 7, 3, 24.5, s.sampleReport.StartTime, s.sampleReport.EndTime)
	s.Nil(err)
}

func (s *ReportServiceTestSuite) TestSendEmailError() {
	s.logger.EXPECT().Error("failed to send email", gomock.Any()).Times(1)
	err := s.reportService.SendEmail(s.ctx, "recipient@example.com", 10, 7, 3, 24.5, s.sampleReport.StartTime, s.sampleReport.EndTime)
	s.Error(err)
}

func (s *ReportServiceTestSuite) TestSendEmailTemplateNotFound() {
	os.Remove("html/email.html")
	s.logger.EXPECT().Error("failed to read email template", gomock.Any()).Times(1)
	err := s.reportService.SendEmail(s.ctx, "recipient@example.com", 10, 7, 3, 24.5, s.sampleReport.StartTime, s.sampleReport.EndTime)
	s.Error(err)
}

func (s *ReportServiceTestSuite) TestSendEmailInvalidTemplate() {
	invalidTemplate := `{{invalid template syntax`
	err := os.WriteFile("html/email.html", []byte(invalidTemplate), 0644)
	s.NoError(err)

	s.logger.EXPECT().Error("failed to parse template", gomock.Any()).Times(1)
	err = s.reportService.SendEmail(s.ctx, "recipient@example.com", 10, 7, 3, 24.5, s.sampleReport.StartTime, s.sampleReport.EndTime)
	s.Error(err)
}

func (s *ReportServiceTestSuite) TestSendEmailTemplateExecutionError() {
	invalidTemplate := `<html><body>{{.NonExistentField}}</body></html>`
	err := os.WriteFile("html/email.html", []byte(invalidTemplate), 0644)
	s.NoError(err)

	s.logger.EXPECT().Error("failed to execute template", gomock.Any()).Times(1)
	err = s.reportService.SendEmail(s.ctx, "recipient@example.com", 10, 7, 3, 24.5, s.sampleReport.StartTime, s.sampleReport.EndTime)
	s.Error(err)
}

func (s *ReportServiceTestSuite) TestCalculateReportStatistic() {
	containers := []*entities.Container{
		{ContainerId: "container1", Status: entities.ContainerOn},
		{ContainerId: "container2", Status: entities.ContainerOff},
		{ContainerId: "container3", Status: entities.ContainerOn},
	}

	statusList := map[string][]dto.EsStatus{
		"container1": {
			{ContainerId: "container1", Status: entities.ContainerOn, LastUpdated: time.Now().Add(-2 * time.Hour)},
		},
		"container2": {
			{ContainerId: "container2", Status: entities.ContainerOff, LastUpdated: time.Now().Add(-1 * time.Hour)},
		},
		"container3": {
			{ContainerId: "container3", Status: entities.ContainerOn, LastUpdated: time.Now().Add(-3 * time.Hour)},
		},
	}

	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now()

	onCount, offCount, totalUptime, err := s.reportService.CalculateReportStatistic(containers, statusList, startTime, endTime)

	s.NoError(err)
	s.Equal(2, onCount)
	s.Equal(1, offCount)
	s.Greater(totalUptime, 0.0)
}

func (s *ReportServiceTestSuite) TestCalculateReportStatisticInvalidDateRange() {
	containers := []*entities.Container{}
	statusList := map[string][]dto.EsStatus{}

	startTime := time.Now()
	endTime := time.Now().Add(-24 * time.Hour)

	s.logger.EXPECT().Error("failed to calculate report statistic", gomock.Any()).Times(1)

	onCount, offCount, totalUptime, err := s.reportService.CalculateReportStatistic(containers, statusList, startTime, endTime)

	s.Error(err)
	s.Contains(err.Error(), "invalid date range")
	s.Equal(0, onCount)
	s.Equal(0, offCount)
	s.Equal(0.0, totalUptime)
}

func (s *ReportServiceTestSuite) TestCalculateReportStatisticFutureEndTime() {
	containers := []*entities.Container{
		{ContainerId: "container1", Status: entities.ContainerOn},
	}

	statusList := map[string][]dto.EsStatus{
		"container1": {
			{ContainerId: "container1", Status: entities.ContainerOn, LastUpdated: time.Now().Add(-1 * time.Hour)},
		},
	}

	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now().Add(24 * time.Hour)

	onCount, offCount, totalUptime, err := s.reportService.CalculateReportStatistic(containers, statusList, startTime, endTime)

	s.NoError(err)
	s.Equal(1, onCount)
	s.Equal(0, offCount)
	s.Greater(totalUptime, 0.0)
}
