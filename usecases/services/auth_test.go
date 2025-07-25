package services

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"github.com/vnFuhung2903/vcs-sms/entities"
	"github.com/vnFuhung2903/vcs-sms/mocks/interfaces"
	"github.com/vnFuhung2903/vcs-sms/mocks/logger"
	"github.com/vnFuhung2903/vcs-sms/mocks/repositories"
	"github.com/vnFuhung2903/vcs-sms/pkg/env"
	"golang.org/x/crypto/bcrypt"
)

type AuthServiceSuite struct {
	suite.Suite
	ctrl        *gomock.Controller
	authService IAuthService
	mockRepo    *repositories.MockIUserRepository
	mockRedis   *interfaces.MockIRedisClient
	logger      *logger.MockILogger
	ctx         context.Context
	testSecret  string
}

func (s *AuthServiceSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockRepo = repositories.NewMockIUserRepository(s.ctrl)
	s.logger = logger.NewMockILogger(s.ctrl)

	s.testSecret = "test-secret-key"
	authEnv := env.AuthEnv{
		JWTSecret: s.testSecret,
	}

	s.authService = NewAuthService(s.mockRepo, s.mockRedis, s.logger, authEnv)
}

func (s *AuthServiceSuite) TearDownTest() {
	s.ctrl.Finish()
}

func TestAuthServiceSuite(t *testing.T) {
	suite.Run(t, new(AuthServiceSuite))
}

func (s *AuthServiceSuite) TestRegister() {
	username := "testuser"
	password := "password123"
	email := "test@example.com"
	role := entities.Developer
	scopes := int64(7)

	expected := &entities.User{
		ID:       "test-id",
		Username: username,
		Email:    email,
		Role:     role,
		Scopes:   scopes,
	}

	s.mockRepo.EXPECT().Create(username, gomock.Any(), email, role, scopes).Return(expected, nil)
	s.logger.EXPECT().Info("new user registered successfully").Times(1)

	result, err := s.authService.Register(username, password, email, role, scopes)
	s.NoError(err)
	s.Equal(expected, result)
}

func (s *AuthServiceSuite) TestRegisterInvalidEmail() {
	username := "testuser"
	password := "password123"
	email := "invalid-email"
	role := entities.Developer
	scopes := int64(7)

	s.logger.EXPECT().Error("failed to parse email", gomock.Any()).Times(1)

	result, err := s.authService.Register(username, password, email, role, scopes)
	s.Error(err)
	s.Nil(result)
}

func (s *AuthServiceSuite) TestRegisterError() {
	username := "testuser"
	password := "password123"
	email := "test@example.com"
	role := entities.Developer
	scopes := int64(7)

	s.mockRepo.EXPECT().Create(username, gomock.Any(), email, role, scopes).Return(nil, errors.New("db error"))
	s.logger.EXPECT().Error("failed to create user", gomock.Any()).Times(1)

	result, err := s.authService.Register(username, password, email, role, scopes)
	s.ErrorContains(err, "db error")
	s.Nil(result)
}

func (s *AuthServiceSuite) TestLoginWithUsername() {
	username := "testuser"
	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	expected := &entities.User{
		ID:       "test-id",
		Username: username,
		Hash:     string(hashedPassword),
	}

	s.mockRepo.EXPECT().FindByName(username).Return(expected, nil)
	s.logger.EXPECT().Info("user logged in successfully").Times(1)

	accessToken, err := s.authService.Login(s.ctx, username, password)
	s.NoError(err)
	s.NotEqual("", accessToken)
}

func (s *AuthServiceSuite) TestLoginWithEmail() {
	email := "test@example.com"
	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	expected := &entities.User{
		ID:    "test-id",
		Email: email,
		Hash:  string(hashedPassword),
	}

	s.mockRepo.EXPECT().FindByEmail(email).Return(expected, nil)
	s.logger.EXPECT().Info("user logged in successfully").Times(1)

	accessToken, err := s.authService.Login(s.ctx, email, password)
	s.NoError(err)
	s.NotEqual("", accessToken)
}

func (s *AuthServiceSuite) TestLoginUserNotFoundByUsername() {
	username := "nonexistent"
	password := "password123"

	s.mockRepo.EXPECT().FindByName(username).Return(nil, errors.New("user not found"))
	s.logger.EXPECT().Error("failed to find user by username", gomock.Any()).Times(1)

	accessToken, err := s.authService.Login(s.ctx, username, password)
	s.Equal("", accessToken)
	s.ErrorContains(err, "user not found")
}

func (s *AuthServiceSuite) TestLoginUserNotFoundByEmail() {
	email := "nonexistent@example.com"
	password := "password123"

	s.mockRepo.EXPECT().FindByEmail(email).Return(nil, errors.New("user not found"))
	s.logger.EXPECT().Error("failed to find user by email", gomock.Any()).Times(1)

	accessToken, err := s.authService.Login(s.ctx, email, password)
	s.Equal("", accessToken)
	s.ErrorContains(err, "user not found")
}

func (s *AuthServiceSuite) TestLoginWrongPassword() {
	username := "testuser"
	password := "wrongpassword"
	correctPassword := "correctpassword"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(correctPassword), bcrypt.DefaultCost)

	user := &entities.User{
		ID:       "test-id",
		Username: username,
		Hash:     string(hashedPassword),
	}

	s.mockRepo.EXPECT().FindByName(username).Return(user, nil)
	s.logger.EXPECT().Error("failed to validate password", gomock.Any()).Times(1)

	accessToken, err := s.authService.Login(s.ctx, username, password)
	s.Equal("", accessToken)
	s.Error(err)
}

func (s *AuthServiceSuite) TestUpdatePassword() {
	userId := "test-id"
	currentPassword := "oldpassword"
	newPassword := "newpassword123"

	existingUser := &entities.User{
		ID:       userId,
		Username: "testuser",
	}

	s.mockRepo.EXPECT().FindById(userId).Return(existingUser, nil)
	s.mockRepo.EXPECT().UpdatePassword(existingUser, gomock.Any()).Return(nil)
	s.logger.EXPECT().Info("user's password updated successfully").Times(1)

	err := s.authService.UpdatePassword(userId, currentPassword, newPassword)
	s.NoError(err)
}

func (s *AuthServiceSuite) TestUpdatePasswordUserNotFound() {
	userId := "nonexistent-id"
	currentPassword := "oldpassword"
	newPassword := "newpassword123"

	s.mockRepo.EXPECT().FindById(userId).Return(nil, errors.New("user not found"))
	s.logger.EXPECT().Error("failed to find user by id", gomock.Any()).Times(1)

	err := s.authService.UpdatePassword(userId, currentPassword, newPassword)
	s.ErrorContains(err, "user not found")
}

func (s *AuthServiceSuite) TestUpdatePasswordError() {
	userId := "test-id"
	currentPassword := "oldpassword"
	newPassword := "newpassword123"

	existingUser := &entities.User{
		ID:       userId,
		Username: "testuser",
	}

	s.mockRepo.EXPECT().FindById(userId).Return(existingUser, nil)
	s.mockRepo.EXPECT().UpdatePassword(existingUser, gomock.Any()).Return(errors.New("update failed"))
	s.logger.EXPECT().Error("failed to update user's password", gomock.Any()).Times(1)

	err := s.authService.UpdatePassword(userId, currentPassword, newPassword)
	s.ErrorContains(err, "update failed")
}
