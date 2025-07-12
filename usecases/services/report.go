package services

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"os"
	"strconv"
	"time"

	"github.com/vnFuhung2903/vcs-sms/dto"
	"github.com/vnFuhung2903/vcs-sms/pkg/logger"
	"go.uber.org/zap"
	"gopkg.in/gomail.v2"
)

var SMTPHost = os.Getenv("SMPT_HOST")
var SMTPPort = os.Getenv("SMPT_PORT")
var mailUsername = os.Getenv("MAIL_USERNAME")
var mailPassword = os.Getenv("MAIL_PASSWORD")

type IReportService interface {
	SendEmail(ctx context.Context, report *dto.ReportResponse, to string, startTime time.Time, endTime time.Time) error
}

type ReportService struct {
	logger logger.ILogger
}

func NewReportService(logger logger.ILogger) IReportService {
	return &ReportService{
		logger: logger,
	}
}

func (s *ReportService) SendEmail(ctx context.Context, report *dto.ReportResponse, to string, startTime time.Time, endTime time.Time) error {
	emailTemplate, err := os.ReadFile("html/email.html")
	if err != nil {
		s.logger.Error("failed to read email template: %w", zap.Error(err))
		return err
	}

	temp, err := template.New("report").Parse(string(emailTemplate))
	if err != nil {
		s.logger.Error("failed to parse template: %w", zap.Error(err))
		return err
	}

	var buf bytes.Buffer
	if err := temp.Execute(&buf, report); err != nil {
		s.logger.Error("failed to execute template: %w", zap.Error(err))
		return err
	}

	msg := fmt.Sprintf("Server Management System Report from %s to %s", startTime.Format(time.RFC822), endTime.Format(time.RFC822))

	message := gomail.NewMessage()
	message.SetHeader("From", mailUsername)
	message.SetHeader("To", to)
	message.SetHeader("Subject", msg)
	message.SetBody("text/html", buf.String())

	port, err := strconv.Atoi(SMTPPort)
	if err != nil {
		s.logger.Error("failed to parse port: %w", zap.Error(err))
		return err
	}

	dial := gomail.NewDialer(
		SMTPHost,
		port,
		mailUsername,
		mailPassword,
	)

	if err := dial.DialAndSend(message); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	s.logger.Info("Report sent successfully", zap.String("emailTo", to), zap.String("subject", msg))
	return nil
}
