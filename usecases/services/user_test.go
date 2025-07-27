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
	"github.com/vnFuhung2903/vcs-sms/utils"
)

type UserServiceSuite struct {
	suite.Suite
	ctrl        *gomock.Controller
	userService IUserService
	mockRepo    *repositories.MockIUserRepository
	mockRedis   *interfaces.MockIRedisClient
	logger      *logger.MockILogger
	ctx         context.Context
}

func (s *UserServiceSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockRepo = repositories.NewMockIUserRepository(s.ctrl)
	s.mockRedis = interfaces.NewMockIRedisClient(s.ctrl)
	s.logger = logger.NewMockILogger(s.ctrl)
	s.userService = NewUserService(s.mockRepo, s.mockRedis, s.logger)
	s.ctx = context.Background()
}

func (s *UserServiceSuite) TearDownTest() {
	s.ctrl.Finish()
}

func TestUserServiceSuite(t *testing.T) {
	suite.Run(t, new(UserServiceSuite))
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
	s.mockRedis.EXPECT().Del(s.ctx, "refresh:"+userId).Return(nil)
	s.logger.EXPECT().Info("user's role updated successfully").Times(1)

	err := s.userService.UpdateRole(s.ctx, userId, newRole)
	s.NoError(err)
}

func (s *UserServiceSuite) TestUpdateRoleUserNotFound() {
	userId := "nonexistent-id"
	newRole := entities.Manager

	s.mockRepo.EXPECT().FindById(userId).Return(nil, errors.New("user not found"))
	s.logger.EXPECT().Error("failed to find user by id", gomock.Any()).Times(1)

	err := s.userService.UpdateRole(s.ctx, userId, newRole)
	s.ErrorContains(err, "user not found")
}

func (s *UserServiceSuite) TestUpdateRoleRepoError() {
	userId := "test-id"
	newRole := entities.Manager

	existingUser := &entities.User{
		ID:   userId,
		Role: entities.Developer,
	}

	s.mockRepo.EXPECT().FindById(userId).Return(existingUser, nil)
	s.mockRepo.EXPECT().UpdateRole(existingUser, newRole).Return(errors.New("update failed"))
	s.logger.EXPECT().Error("failed to update user's role", gomock.Any()).Times(1)

	err := s.userService.UpdateRole(s.ctx, userId, newRole)
	s.ErrorContains(err, "update failed")
}

func (s *UserServiceSuite) TestUpdateRoleRedisError() {
	userId := "test-id"
	newRole := entities.Manager

	existingUser := &entities.User{
		ID:   userId,
		Role: entities.Developer,
	}

	s.mockRepo.EXPECT().FindById(userId).Return(existingUser, nil)
	s.mockRepo.EXPECT().UpdateRole(existingUser, newRole).Return(nil)
	s.mockRedis.EXPECT().Del(s.ctx, "refresh:"+userId).Return(errors.New("redis error"))
	s.logger.EXPECT().Error("failed to delete refresh token in redis", gomock.Any()).Times(1)

	err := s.userService.UpdateRole(s.ctx, userId, newRole)
	s.ErrorContains(err, "redis error")
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
	s.mockRedis.EXPECT().Del(s.ctx, "refresh:"+userId).Return(nil)
	s.logger.EXPECT().Info("user's scopes updated successfully").Times(1)

	err := s.userService.UpdateScope(s.ctx, userId, scopes, isAdded)
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
	s.mockRedis.EXPECT().Del(s.ctx, "refresh:"+userId).Return(nil)
	s.logger.EXPECT().Info("user's scopes updated successfully").Times(1)

	err := s.userService.UpdateScope(s.ctx, userId, scopes, isAdded)
	s.NoError(err)
}

func (s *UserServiceSuite) TestUpdateScopeUserNotFound() {
	userId := "nonexistent-id"
	scopes := []string{"user:modify"}
	isAdded := true

	s.mockRepo.EXPECT().FindById(userId).Return(nil, errors.New("user not found"))
	s.logger.EXPECT().Error("failed to find user by id", gomock.Any()).Times(1)

	err := s.userService.UpdateScope(s.ctx, userId, scopes, isAdded)
	s.ErrorContains(err, "user not found")
}

func (s *UserServiceSuite) TestUpdateScopeRepoError() {
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

	err := s.userService.UpdateScope(s.ctx, userId, scopes, isAdded)
	s.ErrorContains(err, "update failed")
}

func (s *UserServiceSuite) TestUpdateScopeRedisError() {
	userId := "test-id"
	scopes := []string{"user:modify"}
	isAdded := true

	existingUser := &entities.User{
		ID:     userId,
		Scopes: 0,
	}

	expectedScopeHashmap := utils.ScopesToHashMap(scopes)

	s.mockRepo.EXPECT().FindById(userId).Return(existingUser, nil)
	s.mockRepo.EXPECT().UpdateScope(existingUser, expectedScopeHashmap).Return(nil)
	s.mockRedis.EXPECT().Del(s.ctx, "refresh:"+userId).Return(errors.New("redis error"))
	s.logger.EXPECT().Error("failed to delete refresh token in redis", gomock.Any()).Times(1)

	err := s.userService.UpdateScope(s.ctx, userId, scopes, isAdded)
	s.ErrorContains(err, "redis error")
}

func (s *UserServiceSuite) TestDelete() {
	userId := "test-id"

	s.mockRepo.EXPECT().Delete(userId).Return(nil)
	s.mockRedis.EXPECT().Del(s.ctx, "refresh:"+userId).Return(nil)
	s.logger.EXPECT().Info("user deleted successfully").Times(1)

	err := s.userService.Delete(s.ctx, userId)
	s.NoError(err)
}

func (s *UserServiceSuite) TestDeleteRepoError() {
	userId := "test-id"

	s.mockRepo.EXPECT().Delete(userId).Return(errors.New("delete failed"))
	s.logger.EXPECT().Error("failed to delete user", gomock.Any()).Times(1)

	err := s.userService.Delete(s.ctx, userId)
	s.ErrorContains(err, "delete failed")
}

func (s *UserServiceSuite) TestDeleteRedisError() {
	userId := "test-id"

	s.mockRepo.EXPECT().Delete(userId).Return(nil)
	s.mockRedis.EXPECT().Del(s.ctx, "refresh:"+userId).Return(errors.New("redis error"))
	s.logger.EXPECT().Error("failed to delete refresh token in redis", gomock.Any()).Times(1)

	err := s.userService.Delete(s.ctx, userId)
	s.ErrorContains(err, "redis error")
}
