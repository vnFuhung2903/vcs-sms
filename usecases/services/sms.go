package services

import (
	"context"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/vnFuhung2903/vcs-sms/pkg/docker"
	"github.com/vnFuhung2903/vcs-sms/pkg/logger"
	"github.com/vnFuhung2903/vcs-sms/usecases/repositories"
)

type ISMSService interface {
	Report(ctx context.Context) error
}

type SMSService struct {
	containerRepo repositories.IContainerRepository
	dockerClient  docker.IDockerClient
	es            *elasticsearch.Client
	logger        logger.ILogger
}

func NewSMSService(repo repositories.IContainerRepository, dockerClient docker.IDockerClient, es *elasticsearch.Client, logger logger.ILogger) ISMSService {
	return &SMSService{
		containerRepo: repo,
		dockerClient:  dockerClient,
		es:            es,
		logger:        logger,
	}
}

func (s *SMSService) Report(ctx context.Context) error {
	return nil
}
