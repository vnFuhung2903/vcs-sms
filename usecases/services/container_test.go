package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"github.com/vnFuhung2903/vcs-sms/dto"
	"github.com/vnFuhung2903/vcs-sms/entities"
	"github.com/vnFuhung2903/vcs-sms/mocks/logger"
	"github.com/vnFuhung2903/vcs-sms/mocks/repositories"
)

// mockTx is a mock transaction for testing
type mockTx struct {
	*gorm.DB
	committed   bool
	rolled      bool
	commitError error
}

func newMockTx(commitError error) *mockTx {
	db := &gorm.DB{}
	if commitError != nil {
		db.Error = commitError
	}

	tx := &mockTx{
		DB:          db,
		commitError: commitError,
	}

	return tx
}

func (m *mockTx) Commit() *gorm.DB {
	m.committed = true
	if m.commitError != nil {
		m.DB.Error = m.commitError
		return m.DB
	}
	m.DB.Error = nil
	return m.DB
}

func (m *mockTx) Rollback() *gorm.DB {
	m.rolled = true
	return m.DB
}

type ContainerServiceSuite struct {
	suite.Suite
	ctrl         *gomock.Controller
	mockRepo     *repositories.MockIContainerRepository
	logger       *logger.MockILogger
	containerSvc IContainerService
	ctx          context.Context
}

func (s *ContainerServiceSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockRepo = repositories.NewMockIContainerRepository(s.ctrl)
	s.logger = logger.NewMockILogger(s.ctrl)
	s.containerSvc = NewContainerService(s.mockRepo, s.logger)
	s.ctx = context.Background()
}

func (s *ContainerServiceSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *ContainerServiceSuite) TestCreate() {
	expected := &entities.Container{ContainerId: "abc123"}

	s.mockRepo.EXPECT().Create("abc123", "container", entities.ContainerStatus("ON"), "127.0.0.1").Return(expected, nil)

	s.logger.EXPECT().Info(gomock.Any(), gomock.Any()).Times(1)

	result, err := s.containerSvc.Create(s.ctx, "abc123", "container", entities.ContainerStatus("ON"), "127.0.0.1")
	s.NoError(err)
	s.Equal(expected, result)
}

func (s *ContainerServiceSuite) TestCreateError() {
	s.mockRepo.EXPECT().Create("abc123", "container", entities.ContainerStatus("ON"), "127.0.0.1").Return(nil, errors.New("db error"))

	s.logger.EXPECT().Error(gomock.Any(), gomock.Any()).Times(1)

	result, err := s.containerSvc.Create(s.ctx, "abc123", "container", entities.ContainerStatus("ON"), "127.0.0.1")
	s.ErrorContains(err, "db error")
	s.Nil(result)
}

func (s *ContainerServiceSuite) TestCreateWithDifferentStatuses() {
	testCases := []struct {
		name   string
		status entities.ContainerStatus
	}{
		{"ON status", entities.ContainerOn},
		{"OFF status", entities.ContainerOff},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			expected := &entities.Container{ContainerId: "test123", Status: tc.status}

			s.mockRepo.EXPECT().Create("test123", "test-container", tc.status, "192.168.1.1").Return(expected, nil)
			s.logger.EXPECT().Info(gomock.Any(), gomock.Any()).Times(1)

			result, err := s.containerSvc.Create(s.ctx, "test123", "test-container", tc.status, "192.168.1.1")
			s.NoError(err)
			s.Equal(expected, result)
		})
	}
}

func (s *ContainerServiceSuite) TestCreateWithInvalidStatus() {
	invalidStatus := entities.ContainerStatus("INVALID")
	expected := &entities.Container{
		ContainerId:   "test123",
		ContainerName: "test-container",
		Status:        invalidStatus,
		Ipv4:          "192.168.1.1",
	}

	s.mockRepo.EXPECT().Create("test123", "test-container", invalidStatus, "192.168.1.1").Return(expected, nil)
	s.logger.EXPECT().Info(gomock.Any(), gomock.Any()).Times(1)

	result, err := s.containerSvc.Create(s.ctx, "test123", "test-container", invalidStatus, "192.168.1.1")
	s.NoError(err) // Service doesn't validate status
	s.Equal(expected, result)
}

func (s *ContainerServiceSuite) TestCreateWithEmptyFields() {
	testCases := []struct {
		name          string
		containerId   string
		containerName string
		ipv4          string
		expectError   bool
	}{
		{"Valid inputs", "container123", "test-container", "192.168.1.1", false},
		{"Empty container ID", "", "test-container", "192.168.1.1", false},
		{"Empty container name", "container123", "", "192.168.1.1", false},
		{"Empty IPv4", "container123", "test-container", "", false},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			expected := &entities.Container{
				ContainerId:   tc.containerId,
				ContainerName: tc.containerName,
				Status:        entities.ContainerOn,
				Ipv4:          tc.ipv4,
			}

			if !tc.expectError {
				s.mockRepo.EXPECT().Create(tc.containerId, tc.containerName, entities.ContainerOn, tc.ipv4).Return(expected, nil)
				s.logger.EXPECT().Info(gomock.Any(), gomock.Any()).Times(1)
			}

			result, err := s.containerSvc.Create(s.ctx, tc.containerId, tc.containerName, entities.ContainerOn, tc.ipv4)
			if tc.expectError {
				s.Error(err)
				s.Nil(result)
			} else {
				s.NoError(err)
				s.Equal(expected, result)
			}
		})
	}
}

func (s *ContainerServiceSuite) TestViewWithEdgeCases() {
	testCases := []struct {
		name        string
		from        int
		to          int
		expectError bool
		expectLimit int
	}{
		{"Normal range", 1, 10, false, 10},
		{"Zero from", 0, 10, true, -1},
		{"Negative from", -1, 10, true, -1},
		{"To less than from", 10, 5, false, -1},
		{"Large range", 1, 1000000, false, 1000000},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			filter := dto.ContainerFilter{}
			sort := dto.ContainerSort{}
			expected := []*entities.Container{{ContainerId: "test"}}

			if !tc.expectError {
				s.mockRepo.EXPECT().View(filter, tc.from, tc.expectLimit, sort).Return(expected, int64(1), nil)
				s.logger.EXPECT().Info(gomock.Any(), gomock.Any()).Times(1)
			} else {
				s.logger.EXPECT().Error(gomock.Any(), gomock.Any()).Times(1)
			}

			result, total, err := s.containerSvc.View(s.ctx, filter, tc.from, tc.to, sort)
			if tc.expectError {
				s.Error(err)
				s.Nil(result)
				s.Equal(int64(0), total)
			} else {
				s.NoError(err)
				s.Equal(expected, result)
				s.Equal(int64(1), total)
			}
		})
	}
}

func (s *ContainerServiceSuite) TestViewInvalidRange() {
	s.logger.EXPECT().Error(gomock.Any(), gomock.Any()).Times(1)
	_, _, err := s.containerSvc.View(s.ctx, dto.ContainerFilter{}, 0, 10, dto.ContainerSort{})
	s.ErrorContains(err, "invalid range")
}

func (s *ContainerServiceSuite) TestView() {
	filter := dto.ContainerFilter{}
	sort := dto.ContainerSort{}
	expected := []*entities.Container{{ContainerId: "abc"}}

	s.mockRepo.EXPECT().View(filter, 1, 10, sort).Return(expected, int64(1), nil)

	s.logger.EXPECT().Info(gomock.Any(), gomock.Any()).Times(1)

	result, total, err := s.containerSvc.View(s.ctx, filter, 1, 10, sort)
	s.NoError(err)
	s.Equal(int64(1), total)
	s.Equal(expected, result)
}

func (s *ContainerServiceSuite) TestViewError() {
	s.logger.EXPECT().Error(gomock.Any(), gomock.Any()).Times(1)

	s.mockRepo.EXPECT().View(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, int64(0), errors.New("db error"))

	_, _, err := s.containerSvc.View(s.ctx, dto.ContainerFilter{}, 1, 10, dto.ContainerSort{})
	s.ErrorContains(err, "db error")
}

func (s *ContainerServiceSuite) TestViewToRangeCalculation() {
	filter := dto.ContainerFilter{}
	sort := dto.ContainerSort{}
	expected := []*entities.Container{{ContainerId: "abc"}}

	s.mockRepo.EXPECT().View(filter, 5, -1, sort).Return(expected, int64(1), nil)
	s.logger.EXPECT().Info(gomock.Any(), gomock.Any()).Times(1)

	result, total, err := s.containerSvc.View(s.ctx, filter, 5, 3, sort)
	s.NoError(err)
	s.Equal(int64(1), total)
	s.Equal(expected, result)
}

func (s *ContainerServiceSuite) TestViewWithFiltersAndSort() {
	filter := dto.ContainerFilter{
		ContainerId:   "test",
		Status:        "ON",
		ContainerName: "container",
		Ipv4:          "192.168.1.1",
	}
	sort := dto.ContainerSort{
		Field: "created_at",
		Sort:  "desc",
	}
	expected := []*entities.Container{
		{ContainerId: "test123", Status: "ON"},
		{ContainerId: "test456", Status: "ON"},
	}

	s.mockRepo.EXPECT().View(filter, 1, 10, sort).Return(expected, int64(2), nil)
	s.logger.EXPECT().Info(gomock.Any(), gomock.Any()).Times(1)

	result, total, err := s.containerSvc.View(s.ctx, filter, 1, 10, sort)
	s.NoError(err)
	s.Equal(int64(2), total)
	s.Equal(expected, result)
}

func (s *ContainerServiceSuite) TestViewWithComplexFilters() {
	complexFilter := dto.ContainerFilter{
		ContainerId:   "partial",
		Status:        entities.ContainerOn,
		ContainerName: "test",
		Ipv4:          "192.168",
	}
	sort := dto.ContainerSort{
		Field: "updated_at",
		Sort:  "asc",
	}
	expected := []*entities.Container{
		{ContainerId: "partial-match-1", Status: entities.ContainerOn},
		{ContainerId: "partial-match-2", Status: entities.ContainerOn},
	}

	s.mockRepo.EXPECT().View(complexFilter, 1, 50, sort).Return(expected, int64(2), nil)
	s.logger.EXPECT().Info(gomock.Any(), gomock.Any()).Times(1)

	result, total, err := s.containerSvc.View(s.ctx, complexFilter, 1, 50, sort)
	s.NoError(err)
	s.Equal(int64(2), total)
	s.Equal(expected, result)
}

func (s *ContainerServiceSuite) TestViewPaginationEdgeCases() {
	testCases := []struct {
		name     string
		from     int
		to       int
		expected int
	}{
		{"Single item", 1, 1, 1},
		{"To equals from", 5, 5, 1},
		{"Large page", 1, 100, 100},
		{"Reverse range", 10, 1, -1},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			filter := dto.ContainerFilter{}
			sort := dto.ContainerSort{}
			expected := []*entities.Container{{ContainerId: "test"}}

			s.mockRepo.EXPECT().View(filter, tc.from, tc.expected, sort).Return(expected, int64(1), nil)
			s.logger.EXPECT().Info(gomock.Any(), gomock.Any()).Times(1)

			result, total, err := s.containerSvc.View(s.ctx, filter, tc.from, tc.to, sort)
			s.NoError(err)
			s.Equal(expected, result)
			s.Equal(int64(1), total)
		})
	}
}

func (s *ContainerServiceSuite) TestUpdateErrorTransaction() {
	s.logger.EXPECT().Error(gomock.Any(), gomock.Any()).Times(1)

	s.mockRepo.EXPECT().BeginTransaction(gomock.Any()).Return(nil, errors.New("begin tx failed"))

	err := s.containerSvc.Update(s.ctx, "id123", dto.ContainerUpdate{})
	s.ErrorContains(err, "begin tx failed")
}

func (s *ContainerServiceSuite) TestUpdateFindByIdFails() {
	tx := newMockTx(nil)

	s.mockRepo.EXPECT().BeginTransaction(gomock.Any()).Return(tx.DB, nil)
	s.mockRepo.EXPECT().WithTransaction(tx.DB).Return(s.mockRepo)

	s.mockRepo.EXPECT().FindById("id123").Return(nil, errors.New("not found"))
	s.logger.EXPECT().Error("failed to find container", gomock.Any()).Times(1)

	s.mockRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
	s.logger.EXPECT().Info("container updated successfully", gomock.Any()).Times(1)

	err := s.containerSvc.Update(s.ctx, "id123", dto.ContainerUpdate{})
	s.NoError(err)
	s.True(tx.committed)
}

func (s *ContainerServiceSuite) TestUpdate() {
	tx := newMockTx(nil)
	container := &entities.Container{
		ContainerId:   "id123",
		ContainerName: "test-container",
		Status:        "ON",
		Ipv4:          "192.168.1.1",
	}
	updateData := dto.ContainerUpdate{Status: "OFF"}

	s.mockRepo.EXPECT().BeginTransaction(gomock.Any()).Return(tx.DB, nil)
	s.mockRepo.EXPECT().WithTransaction(tx.DB).Return(s.mockRepo)
	s.mockRepo.EXPECT().FindById("id123").Return(container, nil)
	s.mockRepo.EXPECT().Update(container, updateData).Return(nil)
	s.logger.EXPECT().Info("container updated successfully", gomock.Any()).Times(1)

	err := s.containerSvc.Update(s.ctx, "id123", updateData)
	s.NoError(err)
	s.True(tx.committed)
}

func (s *ContainerServiceSuite) TestUpdateFails() {
	tx := newMockTx(nil)
	container := &entities.Container{
		ContainerId:   "id123",
		ContainerName: "test-container",
		Status:        "ON",
		Ipv4:          "192.168.1.1",
	}
	updateData := dto.ContainerUpdate{Status: "OFF"}

	s.mockRepo.EXPECT().BeginTransaction(gomock.Any()).Return(tx.DB, nil)
	s.mockRepo.EXPECT().WithTransaction(tx.DB).Return(s.mockRepo)
	s.mockRepo.EXPECT().FindById("id123").Return(container, nil)
	s.mockRepo.EXPECT().Update(container, updateData).Return(errors.New("update failed"))
	s.logger.EXPECT().Error("failed to update container", gomock.Any()).Times(1)
	s.logger.EXPECT().Info("container updated successfully", gomock.Any()).Times(1)

	err := s.containerSvc.Update(s.ctx, "id123", updateData)
	s.NoError(err)
	s.True(tx.committed)
}

func (s *ContainerServiceSuite) TestUpdateCommitFails() {
	commitError := errors.New("commit failed")
	tx := newMockTx(commitError)
	container := &entities.Container{
		ContainerId:   "id123",
		ContainerName: "test-container",
		Status:        "ON",
		Ipv4:          "192.168.1.1",
	}
	updateData := dto.ContainerUpdate{Status: "OFF"}

	s.mockRepo.EXPECT().BeginTransaction(gomock.Any()).Return(tx.DB, nil)
	s.mockRepo.EXPECT().WithTransaction(tx.DB).Return(s.mockRepo)
	s.mockRepo.EXPECT().FindById("id123").Return(container, nil)
	s.mockRepo.EXPECT().Update(container, updateData).Return(nil)
	s.logger.EXPECT().Error("failed to commit transaction", gomock.Any()).Times(1)

	err := s.containerSvc.Update(s.ctx, "id123", updateData)
	s.ErrorContains(err, "commit failed")
	s.True(tx.committed)
}

func (s *ContainerServiceSuite) TestUpdateRollback() {
	tx := newMockTx(nil)

	s.mockRepo.EXPECT().BeginTransaction(gomock.Any()).Return(tx.DB, nil)
	s.mockRepo.EXPECT().WithTransaction(tx.DB).Return(s.mockRepo)
	s.mockRepo.EXPECT().FindById("id123").DoAndReturn(func(string) (*entities.Container, error) {
		panic("test panic")
	})

	s.Panics(func() {
		s.containerSvc.Update(s.ctx, "id123", dto.ContainerUpdate{})
	})
	s.True(tx.rolled)
}

func (s *ContainerServiceSuite) TestDelete() {
	s.mockRepo.EXPECT().Delete("abc123").Return(nil)
	err := s.containerSvc.Delete(s.ctx, "abc123")
	s.NoError(err)
}

func (s *ContainerServiceSuite) TestDeleteError() {
	s.mockRepo.EXPECT().Delete("abc123").Return(errors.New("delete failed"))
	s.logger.EXPECT().Error(gomock.Any(), gomock.Any()).Times(1)
	err := s.containerSvc.Delete(s.ctx, "abc123")
	s.Error(err)
}

func (s *ContainerServiceSuite) TestDeleteWithSpecialContainerIds() {
	testCases := []struct {
		name        string
		containerId string
		repoError   error
		expectError bool
	}{
		{"Normal ID", "container123", nil, false},
		{"ID with spaces", "container 123", nil, false},
		{"ID with special chars", "container-@#$%", nil, false},
		{"Empty ID", "", nil, false},
		{"Database error", "container123", errors.New("constraint violation"), true},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.mockRepo.EXPECT().Delete(tc.containerId).Return(tc.repoError)
			if tc.expectError {
				s.logger.EXPECT().Error(gomock.Any(), gomock.Any()).Times(1)
			}

			err := s.containerSvc.Delete(s.ctx, tc.containerId)
			if tc.expectError {
				s.Error(err)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *ContainerServiceSuite) TestImportInvalidExcelFile() {
	data := []byte("this is not a real Excel file")
	reader := bytes.NewReader(data)

	fakeFile := struct {
		io.Reader
		io.ReaderAt
		io.Seeker
		io.Closer
	}{
		Reader:   reader,
		ReaderAt: reader,
		Seeker:   reader,
		Closer:   io.NopCloser(nil),
	}

	s.logger.EXPECT().Error(gomock.Any(), gomock.Any()).Times(1)

	resp, err := s.containerSvc.Import(s.ctx, fakeFile)
	s.Error(err)
	s.Nil(resp)
}

func (s *ContainerServiceSuite) TestImportReadRowsError() {
	data := []byte("PK\x03\x04")
	reader := bytes.NewReader(data)

	fakeFile := struct {
		io.Reader
		io.ReaderAt
		io.Seeker
		io.Closer
	}{
		Reader:   reader,
		ReaderAt: reader,
		Seeker:   reader,
		Closer:   io.NopCloser(nil),
	}

	s.logger.EXPECT().Error(gomock.Any(), gomock.Any()).Times(1)

	resp, err := s.containerSvc.Import(s.ctx, fakeFile)
	s.Error(err)
	s.Nil(resp)
}

func (s *ContainerServiceSuite) TestExport() {
	filter := dto.ContainerFilter{}
	sort := dto.ContainerSort{}
	from, to := 1, 10
	containers := []*entities.Container{
		{
			ContainerId:   "abc",
			ContainerName: "test",
			Status:        "ON",
			Ipv4:          "192.168.1.1",
			CreatedAt:     time.Now(),
		},
	}

	s.mockRepo.EXPECT().View(filter, from, to-from+1, sort).Return(containers, int64(len(containers)), nil)

	s.logger.EXPECT().Info(gomock.Any(), gomock.Any()).Times(1)

	result, err := s.containerSvc.Export(s.ctx, filter, from, to, sort)
	s.NoError(err)
	s.True(len(result) > 0)
}

func (s *ContainerServiceSuite) TestExportInvalidRange() {
	s.logger.EXPECT().Error(gomock.Any(), gomock.Any()).Times(1)
	_, err := s.containerSvc.Export(s.ctx, dto.ContainerFilter{}, 0, 10, dto.ContainerSort{})
	s.ErrorContains(err, "invalid range")
}

func (s *ContainerServiceSuite) TestExportError() {
	filter := dto.ContainerFilter{}
	sort := dto.ContainerSort{}
	from, to := 1, 5

	s.mockRepo.EXPECT().
		View(filter, from, to-from+1, sort).
		Return(nil, int64(0), errors.New("fetch error"))

	s.logger.EXPECT().Error(gomock.Any(), gomock.Any()).Times(1)

	_, err := s.containerSvc.Export(s.ctx, filter, from, to, sort)
	s.Error(err)
}

func (s *ContainerServiceSuite) TestExportWithLargeRange() {
	filter := dto.ContainerFilter{}
	sort := dto.ContainerSort{}
	from, to := 1, 1000
	containers := make([]*entities.Container, 100)
	for i := 0; i < 100; i++ {
		containers[i] = &entities.Container{
			ContainerId:   fmt.Sprintf("container-%d", i),
			ContainerName: fmt.Sprintf("name-%d", i),
			Status:        "ON",
			Ipv4:          fmt.Sprintf("192.168.1.%d", i+1),
			CreatedAt:     time.Now(),
		}
	}

	s.mockRepo.EXPECT().View(filter, from, to-from+1, sort).Return(containers, int64(len(containers)), nil)
	s.logger.EXPECT().Info(gomock.Any(), gomock.Any()).Times(1)

	result, err := s.containerSvc.Export(s.ctx, filter, from, to, sort)
	s.NoError(err)
	s.True(len(result) > 0)
}

func (s *ContainerServiceSuite) TestExportWithEmptyData() {
	filter := dto.ContainerFilter{}
	sort := dto.ContainerSort{}
	from, to := 1, 10
	containers := []*entities.Container{} // Empty slice

	s.mockRepo.EXPECT().View(filter, from, to-from+1, sort).Return(containers, int64(0), nil)
	s.logger.EXPECT().Info(gomock.Any(), gomock.Any()).Times(1)

	result, err := s.containerSvc.Export(s.ctx, filter, from, to, sort)
	s.NoError(err)
	s.True(len(result) > 0) // Should still create Excel file with headers
}

func (s *ContainerServiceSuite) TestExportWithSpecialCharacters() {
	filter := dto.ContainerFilter{}
	sort := dto.ContainerSort{}
	from, to := 1, 10
	containers := []*entities.Container{
		{
			ContainerId:   "container-with-特殊字符",
			ContainerName: "test with spaces & symbols!@#",
			Status:        "ON",
			Ipv4:          "192.168.1.1",
			CreatedAt:     time.Now(),
		},
	}

	s.mockRepo.EXPECT().View(filter, from, to-from+1, sort).Return(containers, int64(1), nil)
	s.logger.EXPECT().Info(gomock.Any(), gomock.Any()).Times(1)

	result, err := s.containerSvc.Export(s.ctx, filter, from, to, sort)
	s.NoError(err)
	s.True(len(result) > 0)
}

func (s *ContainerServiceSuite) TestExportRangeValidation() {
	testCases := []struct {
		name        string
		from        int
		to          int
		expectError bool
	}{
		{"Valid range", 1, 10, false},
		{"Invalid from", 0, 10, true},
		{"Negative from", -1, 10, true},
		{"To less than from", 10, 5, false},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			filter := dto.ContainerFilter{}
			sort := dto.ContainerSort{}

			if tc.expectError {
				s.logger.EXPECT().Error(gomock.Any(), gomock.Any()).Times(1)
			} else {
				containers := []*entities.Container{{ContainerId: "test", CreatedAt: time.Now()}}
				limit := max(tc.to-tc.from+1, 1)
				s.mockRepo.EXPECT().View(filter, tc.from, limit, sort).Return(containers, int64(1), nil)
				s.logger.EXPECT().Info(gomock.Any(), gomock.Any()).Times(1)
			}

			result, err := s.containerSvc.Export(s.ctx, filter, tc.from, tc.to, sort)
			if tc.expectError {
				s.Error(err)
				s.Nil(result)
			} else {
				s.NoError(err)
				s.NotNil(result)
			}
		})
	}
}

func TestContainerServiceSuite(t *testing.T) {
	suite.Run(t, new(ContainerServiceSuite))
}
