package services

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"

	"github.com/vnFuhung2903/vcs-sms/entities"
	"github.com/vnFuhung2903/vcs-sms/mocks/logger"
	"github.com/vnFuhung2903/vcs-sms/mocks/repositories"
	"github.com/vnFuhung2903/vcs-sms/utils"
)

type UserServiceSuite struct {
	suite.Suite
	ctrl        *gomock.Controller
	userService IUserService
	mockRepo    *repositories.MockIUserRepository
	logger      *logger.MockILogger
}

func (s *UserServiceSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockRepo = repositories.NewMockIUserRepository(s.ctrl)
	s.logger = logger.NewMockILogger(s.ctrl)
	s.userService = NewUserService(s.mockRepo, s.logger)
}

func (s *UserServiceSuite) TearDownTest() {
	s.ctrl.Finish()
}

func TestUserServiceSuite(t *testing.T) {
	suite.Run(t, new(UserServiceSuite))
}

func (s *UserServiceSuite) TestRegister() {
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

	result, err := s.userService.Register(username, password, email, role, scopes)
	s.NoError(err)
	s.Equal(expected, result)
}

func (s *UserServiceSuite) TestRegisterInvalidEmail() {
	username := "testuser"
	password := "password123"
	email := "invalid-email"
	role := entities.Developer
	scopes := int64(7)

	s.logger.EXPECT().Error("failed to parse email", gomock.Any()).Times(1)

	result, err := s.userService.Register(username, password, email, role, scopes)
	s.Error(err)
	s.Nil(result)
}

func (s *UserServiceSuite) TestRegisterError() {
	username := "testuser"
	password := "password123"
	email := "test@example.com"
	role := entities.Developer
	scopes := int64(7)

	s.mockRepo.EXPECT().Create(username, gomock.Any(), email, role, scopes).Return(nil, errors.New("db error"))
	s.logger.EXPECT().Error("failed to create user", gomock.Any()).Times(1)

	result, err := s.userService.Register(username, password, email, role, scopes)
	s.ErrorContains(err, "db error")
	s.Nil(result)
}

func (s *UserServiceSuite) TestLoginWithUsername() {
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

	result, err := s.userService.Login(username, password)
	s.NoError(err)
	s.Equal(expected, result)
}

func (s *UserServiceSuite) TestLoginWithEmail() {
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

	result, err := s.userService.Login(email, password)
	s.NoError(err)
	s.Equal(expected, result)
}

func (s *UserServiceSuite) TestLoginUserNotFoundByUsername() {
	username := "nonexistent"
	password := "password123"

	s.mockRepo.EXPECT().FindByName(username).Return(nil, errors.New("user not found"))
	s.logger.EXPECT().Error("failed to find user by username", gomock.Any()).Times(1)

	result, err := s.userService.Login(username, password)
	s.ErrorContains(err, "user not found")
	s.Nil(result)
}

func (s *UserServiceSuite) TestLoginUserNotFoundByEmail() {
	email := "nonexistent@example.com"
	password := "password123"

	s.mockRepo.EXPECT().FindByEmail(email).Return(nil, errors.New("user not found"))
	s.logger.EXPECT().Error("failed to find user by email", gomock.Any()).Times(1)

	result, err := s.userService.Login(email, password)
	s.ErrorContains(err, "user not found")
	s.Nil(result)
}

func (s *UserServiceSuite) TestLoginWrongPassword() {
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

	result, err := s.userService.Login(username, password)
	s.Error(err)
	s.Nil(result)
}

func (s *UserServiceSuite) TestUpdatePassword() {
	userId := "test-id"
	newPassword := "newpassword123"

	existingUser := &entities.User{
		ID:       userId,
		Username: "testuser",
	}

	s.mockRepo.EXPECT().FindById(userId).Return(existingUser, nil)
	s.mockRepo.EXPECT().UpdatePassword(existingUser, gomock.Any()).Return(nil)
	s.logger.EXPECT().Info("user's password updated successfully").Times(1)

	err := s.userService.UpdatePassword(userId, newPassword)
	s.NoError(err)
}

func (s *UserServiceSuite) TestUpdatePasswordUserNotFound() {
	userId := "nonexistent-id"
	newPassword := "newpassword123"

	s.mockRepo.EXPECT().FindById(userId).Return(nil, errors.New("user not found"))
	s.logger.EXPECT().Error("failed to find user by id", gomock.Any()).Times(1)

	err := s.userService.UpdatePassword(userId, newPassword)
	s.ErrorContains(err, "user not found")
}

func (s *UserServiceSuite) TestUpdatePasswordError() {
	userId := "test-id"
	newPassword := "newpassword123"

	existingUser := &entities.User{
		ID:       userId,
		Username: "testuser",
	}

	s.mockRepo.EXPECT().FindById(userId).Return(existingUser, nil)
	s.mockRepo.EXPECT().UpdatePassword(existingUser, gomock.Any()).Return(errors.New("update failed"))
	s.logger.EXPECT().Error("failed to update user's password", gomock.Any()).Times(1)

	err := s.userService.UpdatePassword(userId, newPassword)
	s.ErrorContains(err, "update failed")
}

func (s *UserServiceSuite) TestUpdateRole() {
	userId := "test-id"
	newRole := entities.Manager

	existingUser := &entities.User{
		ID:   userId,
		Role: entities.Developer,
	}

	s.mockRepo.EXPECT().FindById(userId).Return(existingUser, nil)
	s.mockRepo.EXPECT().UpdateRole(existingUser, newRole).Return(nil)
	s.logger.EXPECT().Info("user's role updated successfully").Times(1)

	err := s.userService.UpdateRole(userId, newRole)
	s.NoError(err)
}

func (s *UserServiceSuite) TestUpdateRoleUserNotFound() {
	userId := "nonexistent-id"
	newRole := entities.Manager

	s.mockRepo.EXPECT().FindById(userId).Return(nil, errors.New("user not found"))
	s.logger.EXPECT().Error("failed to find user by id", gomock.Any()).Times(1)

	err := s.userService.UpdateRole(userId, newRole)
	s.ErrorContains(err, "user not found")
}

func (s *UserServiceSuite) TestUpdateRoleError() {
	userId := "test-id"
	newRole := entities.Manager

	existingUser := &entities.User{
		ID:   userId,
		Role: entities.Developer,
	}

	s.mockRepo.EXPECT().FindById(userId).Return(existingUser, nil)
	s.mockRepo.EXPECT().UpdateRole(existingUser, newRole).Return(errors.New("update failed"))
	s.logger.EXPECT().Error("failed to update user's role", gomock.Any()).Times(1)

	err := s.userService.UpdateRole(userId, newRole)
	s.ErrorContains(err, "update failed")
}

func (s *UserServiceSuite) TestUpdateScopeAdd() {
	userId := "test-id"
	scopes := []string{"user:modify", "container:create"}
	isAdded := true

	existingUser := &entities.User{
		ID:     userId,
		Scopes: 0,
	}

	expectedScopeHashmap := utils.ScopesToHashMap(scopes)

	s.mockRepo.EXPECT().FindById(userId).Return(existingUser, nil)
	s.mockRepo.EXPECT().UpdateScope(existingUser, expectedScopeHashmap).Return(nil)
	s.logger.EXPECT().Info("user's scopes updated successfully").Times(1)

	err := s.userService.UpdateScope(userId, scopes, isAdded)
	s.NoError(err)
}

func (s *UserServiceSuite) TestUpdateScopeRemove() {
	userId := "test-id"
	scopes := []string{"user:modify", "container:create"}
	isAdded := false

	existingUser := &entities.User{
		ID:     userId,
		Scopes: 0,
	}

	scopeHashmap := utils.ScopesToHashMap(scopes)
	expectedScopeHashmap := scopeHashmap ^ ((1 << utils.NumberOfScopes()) - 1)

	s.mockRepo.EXPECT().FindById(userId).Return(existingUser, nil)
	s.mockRepo.EXPECT().UpdateScope(existingUser, expectedScopeHashmap).Return(nil)
	s.logger.EXPECT().Info("user's scopes updated successfully").Times(1)

	err := s.userService.UpdateScope(userId, scopes, isAdded)
	s.NoError(err)
}

func (s *UserServiceSuite) TestUpdateScopeUserNotFound() {
	userId := "nonexistent-id"
	scopes := []string{"user:modify"}
	isAdded := true

	s.mockRepo.EXPECT().FindById(userId).Return(nil, errors.New("user not found"))
	s.logger.EXPECT().Error("failed to find user by id", gomock.Any()).Times(1)

	err := s.userService.UpdateScope(userId, scopes, isAdded)
	s.ErrorContains(err, "user not found")
}

func (s *UserServiceSuite) TestUpdateScopeError() {
	userId := "test-id"
	scopes := []string{"user:modify"}
	isAdded := true

	existingUser := &entities.User{
		ID:     userId,
		Scopes: 0,
	}

	expectedScopeHashmap := utils.ScopesToHashMap(scopes)

	s.mockRepo.EXPECT().FindById(userId).Return(existingUser, nil)
	s.mockRepo.EXPECT().UpdateScope(existingUser, expectedScopeHashmap).Return(errors.New("update failed"))
	s.logger.EXPECT().Error("failed to update user's scopes", gomock.Any()).Times(1)

	err := s.userService.UpdateScope(userId, scopes, isAdded)
	s.ErrorContains(err, "update failed")
}

func (s *UserServiceSuite) TestDelete() {
	userId := "test-id"

	s.mockRepo.EXPECT().Delete(userId).Return(nil)
	s.logger.EXPECT().Info("user deleted successfully").Times(1)

	err := s.userService.Delete(userId)
	s.NoError(err)
}

func (s *UserServiceSuite) TestDeleteError() {
	userId := "test-id"

	s.mockRepo.EXPECT().Delete(userId).Return(errors.New("delete failed"))
	s.logger.EXPECT().Error("failed to delete user", gomock.Any()).Times(1)

	err := s.userService.Delete(userId)
	s.ErrorContains(err, "delete failed")
}
