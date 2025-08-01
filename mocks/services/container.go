// Code generated by MockGen. DO NOT EDIT.
// Source: E:\Code\VCS\vcs-sms\usecases\services\container.go

// Package services is a generated GoMock package.
package services

import (
	context "context"
	multipart "mime/multipart"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	dto "github.com/vnFuhung2903/vcs-sms/dto"
	entities "github.com/vnFuhung2903/vcs-sms/entities"
)

// MockIContainerService is a mock of IContainerService interface.
type MockIContainerService struct {
	ctrl     *gomock.Controller
	recorder *MockIContainerServiceMockRecorder
}

// MockIContainerServiceMockRecorder is the mock recorder for MockIContainerService.
type MockIContainerServiceMockRecorder struct {
	mock *MockIContainerService
}

// NewMockIContainerService creates a new mock instance.
func NewMockIContainerService(ctrl *gomock.Controller) *MockIContainerService {
	mock := &MockIContainerService{ctrl: ctrl}
	mock.recorder = &MockIContainerServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIContainerService) EXPECT() *MockIContainerServiceMockRecorder {
	return m.recorder
}

// Create mocks base method.
func (m *MockIContainerService) Create(ctx context.Context, containerName, imageName string) (*entities.Container, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", ctx, containerName, imageName)
	ret0, _ := ret[0].(*entities.Container)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Create indicates an expected call of Create.
func (mr *MockIContainerServiceMockRecorder) Create(ctx, containerName, imageName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockIContainerService)(nil).Create), ctx, containerName, imageName)
}

// Delete mocks base method.
func (m *MockIContainerService) Delete(ctx context.Context, containerId string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", ctx, containerId)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *MockIContainerServiceMockRecorder) Delete(ctx, containerId interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockIContainerService)(nil).Delete), ctx, containerId)
}

// Export mocks base method.
func (m *MockIContainerService) Export(ctx context.Context, filter dto.ContainerFilter, from, to int, sort dto.ContainerSort) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Export", ctx, filter, from, to, sort)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Export indicates an expected call of Export.
func (mr *MockIContainerServiceMockRecorder) Export(ctx, filter, from, to, sort interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Export", reflect.TypeOf((*MockIContainerService)(nil).Export), ctx, filter, from, to, sort)
}

// Import mocks base method.
func (m *MockIContainerService) Import(ctx context.Context, file multipart.File) (*dto.ImportResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Import", ctx, file)
	ret0, _ := ret[0].(*dto.ImportResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Import indicates an expected call of Import.
func (mr *MockIContainerServiceMockRecorder) Import(ctx, file interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Import", reflect.TypeOf((*MockIContainerService)(nil).Import), ctx, file)
}

// Update mocks base method.
func (m *MockIContainerService) Update(ctx context.Context, containerId string, updateData dto.ContainerUpdate) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", ctx, containerId, updateData)
	ret0, _ := ret[0].(error)
	return ret0
}

// Update indicates an expected call of Update.
func (mr *MockIContainerServiceMockRecorder) Update(ctx, containerId, updateData interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockIContainerService)(nil).Update), ctx, containerId, updateData)
}

// View mocks base method.
func (m *MockIContainerService) View(ctx context.Context, containerFilter dto.ContainerFilter, from, to int, sort dto.ContainerSort) ([]*entities.Container, int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "View", ctx, containerFilter, from, to, sort)
	ret0, _ := ret[0].([]*entities.Container)
	ret1, _ := ret[1].(int64)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// View indicates an expected call of View.
func (mr *MockIContainerServiceMockRecorder) View(ctx, containerFilter, from, to, sort interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "View", reflect.TypeOf((*MockIContainerService)(nil).View), ctx, containerFilter, from, to, sort)
}
