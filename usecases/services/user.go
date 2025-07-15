package services

import (
	"errors"
	"net/mail"

	"github.com/vnFuhung2903/vcs-sms/entities"
	"github.com/vnFuhung2903/vcs-sms/pkg/logger"
	"github.com/vnFuhung2903/vcs-sms/usecases/repositories"
	"github.com/vnFuhung2903/vcs-sms/utils"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type IUserService interface {
	Register(username, password, email string, role entities.UserRole, scopes int64) (*entities.User, error)
	Login(username, password string) (*entities.User, error)
	UpdatePassword(userId, password string) error
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

func (s *userService) Register(username, password, email string, role entities.UserRole, scopes int64) (*entities.User, error) {
	existing, err := s.userRepo.FindByName(username)
	if existing != nil || (err != nil && !errors.Is(err, gorm.ErrRecordNotFound)) {
		if err == nil {
			err = errors.New("username already taken")
		}
		s.logger.Error("failed to register user", zap.Error(err))
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("failed to hash password", zap.Error(err))
		return nil, err
	}

	mail, err := mail.ParseAddress(email)
	if err != nil {
		s.logger.Error("failed to parse email", zap.Error(err))
		return nil, err
	}

	user, err := s.userRepo.Create(username, string(hash), mail.Address, role, scopes)
	if err != nil {
		s.logger.Error("failed to create user", zap.Error(err))
		return nil, err
	}

	s.logger.Info("new user registered successfully")
	return user, nil
}

func (s *userService) Login(username, password string) (*entities.User, error) {
	var user *entities.User
	mail, err := mail.ParseAddress(username)
	if err != nil {
		user, err = s.userRepo.FindByName(username)
		if err != nil {
			s.logger.Error("failed to find user by username", zap.Error(err))
			return nil, err
		}
	} else {
		user, err = s.userRepo.FindByEmail(mail.Address)
		if err != nil {
			s.logger.Error("failed to find user by email", zap.Error(err))
			return nil, err
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Hash), []byte(password)); err != nil {
		s.logger.Error("failed to validate password", zap.Error(err))
		return nil, err
	}
	s.logger.Info("user logged in successfully")
	return user, nil
}

func (s *userService) UpdatePassword(userId, password string) error {
	user, err := s.userRepo.FindById(userId)
	if err != nil {
		s.logger.Error("failed to find user by id", zap.Error(err))
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("failed to hash password", zap.Error(err))
		return err
	}

	if err := s.userRepo.UpdatePassword(user, string(hash)); err != nil {
		s.logger.Error("failed to update user's password", zap.Error(err))
		return err
	}
	s.logger.Info("user's password updated successfully")
	return nil
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
