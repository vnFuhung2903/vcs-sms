package services

import (
	"context"

	"github.com/vnFuhung2903/vcs-sms/entities"
	"github.com/vnFuhung2903/vcs-sms/usecases/repositories"
)

type IServerService interface {
	Create(ctx context.Context, serverID string, serverName string, status entities.ServerStatus, ipv4 string) error
	Filter(ctx context.Context, serverFilter *entities.ServerFilter, from, to int, sortOpt entities.ServerSort) ([]*entities.Server, error)
	Delete(ctx context.Context, serverID string) error
}

type serverService struct {
	serverRepo repositories.IServerRepository
}

func NewServerService(repo repositories.IServerRepository) IServerService {
	return &serverService{
		serverRepo: repo,
	}
}

func (s *serverService) Create(ctx context.Context, serverID string, serverName string, status entities.ServerStatus, ipv4 string) error {
	return nil
}

func (s *serverService) Filter(ctx context.Context, serverFilter *entities.ServerFilter, from int, to int, sortOpt entities.ServerSort) ([]*entities.Server, error) {
	return nil, nil
}

func (s *serverService) Delete(ctx context.Context, serverID string) error {
	err := s.serverRepo.Delete(serverID)
	return err
}
