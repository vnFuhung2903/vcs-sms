// Code generated by MockGen. DO NOT EDIT.
// Source: E:\Code\VCS\vcs-sms\usecases\services\auth.go

// Package services is a generated GoMock package.
package services

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	entities "github.com/vnFuhung2903/vcs-sms/entities"
)

// MockIAuthService is a mock of IAuthService interface.
type MockIAuthService struct {
	ctrl     *gomock.Controller
	recorder *MockIAuthServiceMockRecorder
}

// MockIAuthServiceMockRecorder is the mock recorder for MockIAuthService.
type MockIAuthServiceMockRecorder struct {
	mock *MockIAuthService
}

// NewMockIAuthService creates a new mock instance.
func NewMockIAuthService(ctrl *gomock.Controller) *MockIAuthService {
	mock := &MockIAuthService{ctrl: ctrl}
	mock.recorder = &MockIAuthServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIAuthService) EXPECT() *MockIAuthServiceMockRecorder {
	return m.recorder
}

// Login mocks base method.
func (m *MockIAuthService) Login(ctx context.Context, username, password string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Login", ctx, username, password)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Login indicates an expected call of Login.
func (mr *MockIAuthServiceMockRecorder) Login(ctx, username, password interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Login", reflect.TypeOf((*MockIAuthService)(nil).Login), ctx, username, password)
}

// RefreshAccessToken mocks base method.
func (m *MockIAuthService) RefreshAccessToken(ctx context.Context, userId string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RefreshAccessToken", ctx, userId)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RefreshAccessToken indicates an expected call of RefreshAccessToken.
func (mr *MockIAuthServiceMockRecorder) RefreshAccessToken(ctx, userId interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RefreshAccessToken", reflect.TypeOf((*MockIAuthService)(nil).RefreshAccessToken), ctx, userId)
}

// Register mocks base method.
func (m *MockIAuthService) Register(username, password, email string, role entities.UserRole, scopes int64) (*entities.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Register", username, password, email, role, scopes)
	ret0, _ := ret[0].(*entities.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Register indicates an expected call of Register.
func (mr *MockIAuthServiceMockRecorder) Register(username, password, email, role, scopes interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Register", reflect.TypeOf((*MockIAuthService)(nil).Register), username, password, email, role, scopes)
}

// UpdatePassword mocks base method.
func (m *MockIAuthService) UpdatePassword(ctx context.Context, userId, currentPassword, newPassword string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdatePassword", ctx, userId, currentPassword, newPassword)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdatePassword indicates an expected call of UpdatePassword.
func (mr *MockIAuthServiceMockRecorder) UpdatePassword(ctx, userId, currentPassword, newPassword interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdatePassword", reflect.TypeOf((*MockIAuthService)(nil).UpdatePassword), ctx, userId, currentPassword, newPassword)
}
