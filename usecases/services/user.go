package services

import (
	"github.com/vnFuhung2903/vcs-sms/entities"
	"github.com/vnFuhung2903/vcs-sms/pkg/logger"
	"github.com/vnFuhung2903/vcs-sms/usecases/repositories"
	"github.com/vnFuhung2903/vcs-sms/utils"
	"go.uber.org/zap"
)

type IUserService interface {
	UpdateRole(userId string, role entities.UserRole) error
	UpdateScope(userId string, scopes []string, isAdded bool) error
	Delete(userId string) error
}

type userService struct {
	userRepo repositories.IUserRepository
	logger   logger.ILogger
}

func NewUserService(userRepo repositories.IUserRepository, logger logger.ILogger) IUserService {
	return &userService{
		userRepo: userRepo,
		logger:   logger,
	}
}

func (s *userService) UpdateRole(userId string, role entities.UserRole) error {
	user, err := s.userRepo.FindById(userId)
	if err != nil {
		s.logger.Error("failed to find user by id", zap.Error(err))
		return err
	}
	if err := s.userRepo.UpdateRole(user, role); err != nil {
		s.logger.Error("failed to update user's role", zap.Error(err))
		return err
	}
	s.logger.Info("user's role updated successfully")
	return nil
}

func (s *userService) UpdateScope(userId string, scopes []string, isAdded bool) error {
	user, err := s.userRepo.FindById(userId)
	if err != nil {
		s.logger.Error("failed to find user by id", zap.Error(err))
		return err
	}

	scopeHashmap := utils.ScopesToHashMap(scopes)
	if !isAdded {
		scopeHashmap = scopeHashmap ^ ((1 << utils.NumberOfScopes()) - 1)
	}

	if err := s.userRepo.UpdateScope(user, scopeHashmap); err != nil {
		s.logger.Error("failed to update user's scopes", zap.Error(err))
		return err
	}
	s.logger.Info("user's scopes updated successfully")
	return nil
}

func (s *userService) Delete(userId string) error {
	if err := s.userRepo.Delete(userId); err != nil {
		s.logger.Error("failed to delete user", zap.Error(err))
		return err
	}
	s.logger.Info("user deleted successfully")
	return nil
}
