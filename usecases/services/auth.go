package services

import (
	"time"

	"github.com/go-redis/redis"
	"github.com/vnFuhung2903/vcs-sms/pkg/logger"
	"github.com/vnFuhung2903/vcs-sms/utils"
	"github.com/vnFuhung2903/vcs-sms/utils/middlewares"
	"go.uber.org/zap"
)

type IAuthService interface {
	Setup(id string, username string, scopes int64) error
}

type authService struct {
	redisClient *redis.Client
	logger      logger.ILogger
}

func NewAuthService(redisClient *redis.Client, logger logger.ILogger) IAuthService {
	return &authService{
		redisClient: redisClient,
		logger:      logger,
	}
}

func (s *authService) Setup(id string, username string, scopes int64) error {
	token, err := middlewares.GenerateJWT(id, username, utils.HashMapToScopes(scopes))
	if err != nil {
		s.logger.Error("failed to generate jwt token", zap.Error(err))
		return err
	}

	if err := s.redisClient.Set("token", token, time.Hour*24*7).Err(); err != nil {
		s.logger.Error("failed to set jwt token in redis", zap.Error(err))
		return err
	}
	s.logger.Info("token set up successfully")
	return nil
}
