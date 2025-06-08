package services

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"strings"

	"github.com/vnFuhung2903/vcs-sms/entities"
	"github.com/vnFuhung2903/vcs-sms/pkg/logger"
	"github.com/vnFuhung2903/vcs-sms/usecases/repositories"
	"go.uber.org/zap"
)

type IServerService interface {
	Create(ctx context.Context, serverID string, serverName string, status entities.ServerStatus, ipv4 string) (*entities.Server, error)
	View(ctx context.Context, serverFilter entities.ServerFilter, from int, to int, sortOpt entities.ServerSort) ([]*entities.Server, int64, error)
	Update(ctx context.Context, serverId string, updateData map[string]interface{}) error
	Delete(ctx context.Context, serverID string) error
}

type ServerService struct {
	serverRepo repositories.IServerRepository
	logger     logger.ILogger
}

type ImportResult struct {
	SuccessCount   int      `json:"success_count"`
	SuccessServers []string `json:"success_servers"`
	FailureCount   int      `json:"failure_count"`
	FailedServers  []string `json:"failed_servers"`
}

func NewServerService(repo repositories.IServerRepository, logger logger.ILogger) IServerService {
	return &ServerService{
		serverRepo: repo,
		logger:     logger,
	}
}

func (s *ServerService) Create(ctx context.Context, serverID string, serverName string, status entities.ServerStatus, ipv4 string) (*entities.Server, error) {
	server, err := s.serverRepo.Create(serverID, serverName, status, ipv4)
	if err != nil {
		s.logger.Error("failed to create server", zap.Error(err))
		return nil, err
	}

	s.logger.Info("server created successfully", zap.String("serverID", serverID))
	return server, nil
}

func (s *ServerService) View(ctx context.Context, filter entities.ServerFilter, from int, to int, sort entities.ServerSort) (servers []*entities.Server, total int64, err error) {
	limit := to - from + 1
	if from < 1 || limit < 1 {
		err = fmt.Errorf("invalid range")
		s.logger.Error("failed to view servers", zap.Error(err))
		return nil, 0, err
	}

	servers, total, err = s.serverRepo.View(filter, from, limit, sort)
	if err != nil {
		s.logger.Error("failed to view servers", zap.Error(err))
		return nil, 0, err
	}

	s.logger.Info("servers listed successfully", zap.Int("count", int(total)))
	return servers, total, nil
}

func (s *ServerService) Update(ctx context.Context, serverId string, updateData map[string]interface{}) error {
	tx, err := s.serverRepo.BeginTransaction(ctx)
	if err != nil {
		s.logger.Error("failed to begin transaction", zap.Error(err))
		return err
	}
	defer func() {
		if re := recover(); re != nil {
			tx.Rollback()
		}
	}()
	serverRepo := s.serverRepo.WithTransaction(tx)

	server, err := serverRepo.FindById(serverId)
	if err != nil {
		s.logger.Error("failed to find server", zap.Error(err))
	}

	err = serverRepo.Update(server, updateData)
	if err != nil {
		s.logger.Error("failed to update server", zap.Error(err))
	}

	err = tx.Commit().Error
	if err != nil {
		s.logger.Error("failed to commit transaction", zap.Error(err))
		return err
	}

	s.logger.Info("server updated successfully", zap.String("serverID", serverId))
	return nil
}

func (s *ServerService) Delete(ctx context.Context, serverID string) error {
	err := s.serverRepo.Delete(serverID)
	if err != nil {
		s.logger.Error("failed to delete server", zap.Error(err))
	}
	return err
}

func (s *ServerService) Import(ctx context.Context, file multipart.File) (*ImportResult, error) {
	f, err := excelize.OpenReader(file)
	if err != nil {
		s.logger.Error("failed to open excel file", zap.Error(err))
		return nil, err
	}
	defer f.Close()

	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		s.logger.Error("failed to read rows", zap.Error(err))
		return nil, err
	}

	result := &ImportResult{}
	for i, row := range rows {
		if i == 0 {
			continue
		}
		if len(row) < 4 {
			s.logger.Warn("skipping invalid row", zap.Int("row", i+1))
			continue
		}

		serverID := strings.TrimSpace(row[0])
		serverName := strings.TrimSpace(row[1])
		status := strings.TrimSpace(row[2])
		ipv4 := strings.TrimSpace(row[3])

		if serverID == "" || serverName == "" || ipv4 == "" {
			result.FailureCount++
			result.FailedServers = append(result.FailedServers, serverID)
			continue
		}

		tx, err := s.serverRepo.BeginTransaction(ctx)
		if err != nil {
			s.logger.Error("failed to begin transaction", zap.Error(err))
			return nil, err
		}
		defer func() {
			if re := recover(); re != nil {
				tx.Rollback()
			}
		}()
		serverRepo := s.serverRepo.WithTransaction(tx)

		existed, err := serverRepo.FindById(serverID)
		if err != nil {
			s.logger.Error("failed to find server by id", zap.Error(err))
			return nil, err
		}
		existed, err = serverRepo.FindByName(serverName)
		if err != nil {
			s.logger.Error("failed to find server by name", zap.Error(err))
			return nil, err
		}

		if existed == nil {
			if status == "" {
				status = "OFF"
			}
			_, err = s.Create(ctx, serverID, serverName, entities.ServerStatus(status), ipv4)
			if err != nil {
				result.FailureCount++
				result.FailedServers = append(result.FailedServers, serverID)
				continue
			}

			result.SuccessCount++
			result.SuccessServers = append(result.SuccessServers, serverID)
		}
		err = tx.Commit().Error
		if err != nil {
			s.logger.Error("failed to commit transaction", zap.Error(err))
			return nil, err
		}
	}
	s.logger.Info("servers imported successfully")
	return result, nil
}

func (s *ServerService) Export(ctx context.Context, filter entities.ServerFilter, from int, to int, sort entities.ServerSort) ([]byte, error) {
	limit := to - from + 1
	if from < 1 || limit < 1 {
		err := fmt.Errorf("invalid range")
		s.logger.Error("failed to view servers", zap.Error(err))
		return nil, err
	}

	servers, _, err := s.serverRepo.View(filter, from, limit, sort)
	if err != nil {
		s.logger.Error("failed to fetch servers for export", zap.Error(err))
		return nil, err
	}

	f := excelize.NewFile()
	sheetName := "Servers"
	f.SetSheetName("Sheet1", sheetName)

	headers := []string{"Server ID", "Server Name", "Status", "IPv4", "Created At"}
	for i, h := range headers {
		cell := fmt.Sprintf("%s1", string(rune('A'+i)))
		f.SetCellValue(sheetName, cell, h)
	}

	for idx, server := range servers {
		row := idx + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), server.ServerId)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), server.ServerName)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), server.Status)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), server.Ipv4)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), server.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		s.logger.Error("failed to write excel buffer", zap.Error(err))
		return nil, err
	}
	s.logger.Info("servers exported successfully")
	return buf.Bytes(), nil
}
