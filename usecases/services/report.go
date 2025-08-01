package services

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"os"
	"time"

	"github.com/vnFuhung2903/vcs-sms/dto"
	"github.com/vnFuhung2903/vcs-sms/entities"
	"github.com/vnFuhung2903/vcs-sms/pkg/env"
	"github.com/vnFuhung2903/vcs-sms/pkg/logger"
	"go.uber.org/zap"
	"gopkg.in/gomail.v2"
)

type IReportService interface {
	SendEmail(ctx context.Context, to string, totalCount int, onCount int, offCount int, totalUptime float64, startTime time.Time, endTime time.Time) error
	CalculateReportStatistic(statusList map[string][]dto.EsStatus, overlapStatusList map[string][]dto.EsStatus, startTime time.Time, endTime time.Time) (int, int, float64)
}

type ReportService struct {
	mailUsername string
	mailPassword string
	logger       logger.ILogger
}

func NewReportService(logger logger.ILogger, env env.GomailEnv) IReportService {
	return &ReportService{
		mailUsername: env.MailUsername,
		mailPassword: env.MailPassword,
		logger:       logger,
	}
}

func (s *ReportService) SendEmail(ctx context.Context, to string, totalCount int, onCount int, offCount int, totalUptime float64, startTime time.Time, endTime time.Time) error {
	emailTemplate, err := os.ReadFile("html/email.html")
	if err != nil {
		s.logger.Error("failed to read email template", zap.Error(err))
		return err
	}

	funcMap := template.FuncMap{
		"formatTime": func(t time.Time) string {
			return t.Format("2006-01-02")
		},
	}
	temp, err := template.New("report").Funcs(funcMap).Parse(string(emailTemplate))
	if err != nil {
		s.logger.Error("failed to parse template", zap.Error(err))
		return err
	}

	report := dto.ReportResponse{
		ContainerCount:    totalCount,
		ContainerOnCount:  onCount,
		ContainerOffCount: offCount,
		TotalUptime:       totalUptime,
		StartTime:         startTime,
		EndTime:           endTime,
	}

	var buf bytes.Buffer
	if err := temp.Execute(&buf, report); err != nil {
		s.logger.Error("failed to execute template", zap.Error(err))
		return err
	}

	msg := fmt.Sprintf("Container Management System Report from %s to %s", startTime.Format(time.RFC822), endTime.Format(time.RFC822))

	message := gomail.NewMessage()
	message.SetHeader("From", s.mailUsername)
	message.SetHeader("To", to)
	message.SetHeader("Subject", msg)
	message.SetBody("text/html", buf.String())

	dial := gomail.NewDialer(
		"smtp.gmail.com",
		587,
		s.mailUsername,
		s.mailPassword,
	)

	if err := dial.DialAndSend(message); err != nil {
		s.logger.Error("failed to send email", zap.Error(err))
		return err
	}

	s.logger.Info("Report sent successfully", zap.String("emailTo", to), zap.String("subject", msg))
	return nil
}

func (s *ReportService) CalculateReportStatistic(statusList map[string][]dto.EsStatus, overlapStatusList map[string][]dto.EsStatus, startTime time.Time, endTime time.Time) (int, int, float64) {
	onCount := 0
	offCount := 0
	totalUptime := 0.0
	isOnline := 0

	for containerId, containerStatus := range statusList {
		previousTime := startTime
		for _, status := range containerStatus {
			if status.Status == entities.ContainerOn {
				totalUptime += min(status.LastUpdated.Sub(startTime).Hours(), float64(status.Uptime)/3600)
				isOnline = 1
			} else {
				previousTime = time.Unix(max(previousTime.Unix(), status.LastUpdated.Unix()), 0)
				isOnline = 0
			}
		}

		if len(overlapStatusList[containerId]) > 0 {
			if overlapStatusList[containerId][0].Status == entities.ContainerOn {
				onCount++
				totalUptime += min(endTime.Sub(previousTime).Hours(), float64(overlapStatusList[containerId][0].Uptime)/3600)
			} else {
				offCount++
			}
			continue
		}

		onCount += isOnline
		offCount += 1 - isOnline
	}

	return onCount, offCount, totalUptime
}
