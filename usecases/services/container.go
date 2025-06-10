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
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
)

type IContainerService interface {
	Create(ctx context.Context, containerID string, containerName string, status entities.ContainerStatus, ipv4 string) (*entities.Container, error)
	View(ctx context.Context, containerFilter entities.ContainerFilter, from int, to int, sortOpt entities.ContainerSort) ([]*entities.Container, int64, error)
	Update(ctx context.Context, containerId string, updateData map[string]any) error
	Import(ctx context.Context, file multipart.File) (*ImportResult, error)
	Export(ctx context.Context, filter entities.ContainerFilter, from int, to int, sort entities.ContainerSort) ([]byte, error)
	Delete(ctx context.Context, containerID string) error
}

type ContainerService struct {
	containerRepo repositories.IContainerRepository
	logger        logger.ILogger
}

type ImportResult struct {
	SuccessCount      int      `json:"success_count"`
	SuccessContainers []string `json:"success_containers"`
	FailureCount      int      `json:"failure_count"`
	FailedContainers  []string `json:"failed_containers"`
}

func NewContainerService(repo repositories.IContainerRepository, logger logger.ILogger) IContainerService {
	return &ContainerService{
		containerRepo: repo,
		logger:        logger,
	}
}

func (s *ContainerService) Create(ctx context.Context, containerID string, containerName string, status entities.ContainerStatus, ipv4 string) (*entities.Container, error) {
	container, err := s.containerRepo.Create(containerID, containerName, status, ipv4)
	if err != nil {
		s.logger.Error("failed to create container", zap.Error(err))
		return nil, err
	}

	s.logger.Info("container created successfully", zap.String("containerID", containerID))
	return container, nil
}

func (s *ContainerService) View(ctx context.Context, filter entities.ContainerFilter, from int, to int, sort entities.ContainerSort) ([]*entities.Container, int64, error) {
	limit := to - from + 1
	if from < 1 || limit < 1 {
		err := fmt.Errorf("invalid range")
		s.logger.Error("failed to view containers", zap.Error(err))
		return nil, 0, err
	}

	containers, total, err := s.containerRepo.View(filter, from, limit, sort)
	if err != nil {
		s.logger.Error("failed to view containers", zap.Error(err))
		return nil, 0, err
	}

	s.logger.Info("containers listed successfully", zap.Int("count", int(total)))
	return containers, total, nil
}

func (s *ContainerService) Update(ctx context.Context, containerId string, updateData map[string]any) error {
	tx, err := s.containerRepo.BeginTransaction(ctx)
	if err != nil {
		s.logger.Error("failed to begin transaction", zap.Error(err))
		return err
	}
	commited := false
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
		if !commited {
			tx.Rollback()
		}
	}()
	containerRepo := s.containerRepo.WithTransaction(tx)

	container, err := containerRepo.FindById(containerId)
	if err != nil {
		s.logger.Error("failed to find container", zap.Error(err))
	}

	if err := containerRepo.Update(container, updateData); err != nil {
		s.logger.Error("failed to update container", zap.Error(err))
	}

	if err := tx.Commit().Error; err != nil {
		s.logger.Error("failed to commit transaction", zap.Error(err))
		return err
	}

	s.logger.Info("container updated successfully", zap.String("containerID", containerId))
	commited = true
	return nil
}

func (s *ContainerService) Delete(ctx context.Context, containerID string) error {
	if err := s.containerRepo.Delete(containerID); err != nil {
		s.logger.Error("failed to delete container", zap.Error(err))
	}
	return nil
}

func (s *ContainerService) Import(ctx context.Context, file multipart.File) (*ImportResult, error) {
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

		containerId := strings.TrimSpace(row[0])
		containerName := strings.TrimSpace(row[1])
		status := strings.TrimSpace(row[2])
		ipv4 := strings.TrimSpace(row[3])

		if containerId == "" || containerName == "" || ipv4 == "" {
			result.FailureCount++
			result.FailedContainers = append(result.FailedContainers, containerId)
			continue
		}

		if err := s._importContainer(ctx, containerId, containerName, entities.ContainerStatus(status), ipv4); err != nil {
			result.FailureCount++
			result.FailedContainers = append(result.FailedContainers, containerId)
			continue
		}
		result.SuccessCount++
		result.SuccessContainers = append(result.SuccessContainers, containerId)
	}
	s.logger.Info("containers imported successfully")
	return result, nil
}

func (s *ContainerService) Export(ctx context.Context, filter entities.ContainerFilter, from int, to int, sort entities.ContainerSort) ([]byte, error) {
	limit := to - from + 1
	if from < 1 || limit < 1 {
		err := fmt.Errorf("invalid range")
		s.logger.Error("failed to view containers", zap.Error(err))
		return nil, err
	}

	containers, _, err := s.containerRepo.View(filter, from, limit, sort)
	if err != nil {
		s.logger.Error("failed to fetch containers for export", zap.Error(err))
		return nil, err
	}

	f := excelize.NewFile()
	sheetName := "Containers"
	f.SetSheetName("Sheet1", sheetName)

	headers := []string{"Container ID", "Container Name", "Status", "IPv4", "Created At"}
	for i, h := range headers {
		cell := fmt.Sprintf("%s1", string(rune('A'+i)))
		f.SetCellValue(sheetName, cell, h)
	}

	for idx, container := range containers {
		row := idx + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), container.ContainerId)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), container.ContainerName)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), container.Status)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), container.Ipv4)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), container.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		s.logger.Error("failed to write excel buffer", zap.Error(err))
		return nil, err
	}
	s.logger.Info("containers exported successfully")
	return buf.Bytes(), nil
}

func (s *ContainerService) _importContainer(ctx context.Context, containerId string, containerName string, status entities.ContainerStatus, ipv4 string) error {
	tx, err := s.containerRepo.BeginTransaction(ctx)
	if err != nil {
		s.logger.Error("failed to begin transaction", zap.Error(err))
		return err
	}
	commited := false
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
		if !commited {
			tx.Rollback()
		}
	}()
	containerRepo := s.containerRepo.WithTransaction(tx)

	existed, err := containerRepo.FindById(containerId)
	if err != nil {
		s.logger.Error("failed to find container by id", zap.Error(err))
		return err
	}
	if existed == nil {
		existed, err = containerRepo.FindByName(containerName)
		if err != nil {
			s.logger.Error("failed to find container by name", zap.Error(err))
			return err
		}
	}

	if existed == nil {
		if status == "" {
			status = "OFF"
		}
		_, err = containerRepo.Create(containerId, containerName, entities.ContainerStatus(status), ipv4)
		if err != nil {
			s.logger.Error("failed to create container", zap.Error(err))
			return err
		}
	}

	if err := tx.Commit().Error; err != nil {
		s.logger.Error("failed to commit transaction", zap.Error(err))
		return err
	}
	commited = true
	return nil
}
