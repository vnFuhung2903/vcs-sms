package services

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"

	"github.com/vnFuhung2903/vcs-sms/dto"
	"github.com/vnFuhung2903/vcs-sms/entities"
	"github.com/vnFuhung2903/vcs-sms/mocks/logger"
	"github.com/vnFuhung2903/vcs-sms/mocks/repositories"
)

type ContainerServiceSuite struct {
	suite.Suite
	ctrl             *gomock.Controller
	containerService IContainerService
	mockRepo         *repositories.MockIContainerRepository
	logger           *logger.MockILogger
	ctx              context.Context
}

func (s *ContainerServiceSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockRepo = repositories.NewMockIContainerRepository(s.ctrl)
	s.logger = logger.NewMockILogger(s.ctrl)
	s.containerService = NewContainerService(s.mockRepo, s.logger)
	s.ctx = context.Background()
}

func (s *ContainerServiceSuite) TearDownTest() {
	s.ctrl.Finish()
}

func TestContainerServiceSuite(t *testing.T) {
	suite.Run(t, new(ContainerServiceSuite))
}

func (s *ContainerServiceSuite) TestCreate() {
	expected := &entities.Container{
		ContainerId:   "test",
		ContainerName: "container",
		Status:        entities.ContainerOn,
		Ipv4:          "127.0.0.1",
	}
	s.mockRepo.EXPECT().Create("test", "container", entities.ContainerStatus("ON"), "127.0.0.1").Return(expected, nil)
	s.logger.EXPECT().Info("container created successfully", zap.String("containerId", expected.ContainerId)).Times(1)

	result, err := s.containerService.Create(s.ctx, "test", "container", entities.ContainerStatus("ON"), "127.0.0.1")
	s.NoError(err)
	s.Equal(expected, result)
}

func (s *ContainerServiceSuite) TestCreateError() {
	s.mockRepo.EXPECT().Create("test", "container", entities.ContainerStatus("ON"), "127.0.0.1").Return(nil, errors.New("db error"))
	s.logger.EXPECT().Error("failed to create container", gomock.Any()).Times(1)

	result, err := s.containerService.Create(s.ctx, "test", "container", entities.ContainerStatus("ON"), "127.0.0.1")
	s.ErrorContains(err, "db error")
	s.Nil(result)
}

func (s *ContainerServiceSuite) TestCreateWithInvalidStatus() {
	invalidStatus := entities.ContainerStatus("INVALID")
	s.logger.EXPECT().Error("failed to create container", gomock.Any()).Times(1)

	result, err := s.containerService.Create(s.ctx, "test123", "test-container", invalidStatus, "192.168.1.1")
	s.Error(err)
	s.Nil(result)
}

func (s *ContainerServiceSuite) TestView() {
	filter := dto.ContainerFilter{}
	sort := dto.ContainerSort{Field: "container_id", Sort: "asc"}
	expected := []*entities.Container{{ContainerId: "abc"}}

	s.mockRepo.EXPECT().View(filter, 1, 10, sort).Return(expected, int64(1), nil)
	s.logger.EXPECT().Info("containers listed successfully", gomock.Any()).Times(1)

	result, total, err := s.containerService.View(s.ctx, filter, 1, 10, sort)
	s.NoError(err)
	s.Equal(int64(1), total)
	s.Equal(expected, result)
}

func (s *ContainerServiceSuite) TestViewError() {
	s.mockRepo.EXPECT().View(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, int64(0), errors.New("db error"))
	s.logger.EXPECT().Error("failed to view containers", gomock.Any()).Times(1)

	_, _, err := s.containerService.View(s.ctx, dto.ContainerFilter{}, 1, 10, dto.ContainerSort{})
	s.ErrorContains(err, "db error")
}

func (s *ContainerServiceSuite) TestViewInvalidRange() {
	s.logger.EXPECT().Error("failed to view containers", gomock.Any()).Times(1)
	_, _, err := s.containerService.View(s.ctx, dto.ContainerFilter{}, 0, 10, dto.ContainerSort{})
	s.ErrorContains(err, "invalid range")
}

func (s *ContainerServiceSuite) TestViewInvalidSort() {
	filter := dto.ContainerFilter{}
	inputSort := dto.ContainerSort{Field: "not-valid", Sort: "not-valid"}
	expectedSort := dto.ContainerSort{Field: "container_id", Sort: "desc"}
	expected := []*entities.Container{{ContainerId: "abc"}}

	s.mockRepo.EXPECT().View(filter, 1, 10, expectedSort).Return(expected, int64(1), nil)
	s.logger.EXPECT().Info("containers listed successfully", gomock.Any()).Times(1)

	result, total, err := s.containerService.View(s.ctx, filter, 1, 10, inputSort)
	s.NoError(err)
	s.Equal(int64(1), total)
	s.Equal(expected, result)
}

func (s *ContainerServiceSuite) TestUpdate() {
	updateData := dto.ContainerUpdate{Status: "OFF"}

	s.mockRepo.EXPECT().Update("test", updateData).Return(nil)
	s.logger.EXPECT().Info("container updated successfully", gomock.Any()).Times(1)

	err := s.containerService.Update(s.ctx, "test", updateData)
	s.NoError(err)
}

func (s *ContainerServiceSuite) TestUpdateError() {
	updateData := dto.ContainerUpdate{Status: "OFF"}

	s.mockRepo.EXPECT().Update("test", updateData).Return(errors.New("update failed"))
	s.logger.EXPECT().Error("failed to update container", gomock.Any()).Times(1)

	err := s.containerService.Update(s.ctx, "test", updateData)
	s.ErrorContains(err, "update failed")
}

func (s *ContainerServiceSuite) TestDelete() {
	s.mockRepo.EXPECT().Delete("test").Return(nil)
	s.logger.EXPECT().Info("container deleted successfully", zap.String("containerId", "test")).Times(1)
	err := s.containerService.Delete(s.ctx, "test")
	s.NoError(err)
}

func (s *ContainerServiceSuite) TestDeleteError() {
	s.mockRepo.EXPECT().Delete("test").Return(errors.New("delete failed"))
	s.logger.EXPECT().Error(gomock.Any(), gomock.Any()).Times(1)
	err := s.containerService.Delete(s.ctx, "test")
	s.Error(err)
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

	resp, err := s.containerService.Import(s.ctx, fakeFile)
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

	resp, err := s.containerService.Import(s.ctx, fakeFile)
	s.Error(err)
	s.Nil(resp)
}

func (s *ContainerServiceSuite) TestExport() {
	filter := dto.ContainerFilter{}
	sort := dto.ContainerSort{Field: "container_id", Sort: "asc"}
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
	s.logger.EXPECT().Info("containers exported successfully").Times(1)

	result, err := s.containerService.Export(s.ctx, filter, from, to, sort)
	s.NoError(err)
	s.True(len(result) > 0)
}

func (s *ContainerServiceSuite) TestExportInvalidRange() {
	s.logger.EXPECT().Error("failed to export containers", gomock.Any()).Times(1)
	_, err := s.containerService.Export(s.ctx, dto.ContainerFilter{}, 0, 10, dto.ContainerSort{})
	s.ErrorContains(err, "invalid range")
}

func (s *ContainerServiceSuite) TestExportInvalidSort() {
	filter := dto.ContainerFilter{}
	inputSort := dto.ContainerSort{Field: "not-valid", Sort: "not-valid"}
	expectedSort := dto.ContainerSort{Field: "container_id", Sort: "desc"}
	from, to := 1, 5
	containers := []*entities.Container{
		{
			ContainerId:   "test",
			ContainerName: "container",
			Status:        "ON",
			Ipv4:          "192.168.1.1",
			CreatedAt:     time.Now(),
		},
	}

	s.mockRepo.EXPECT().View(filter, from, to-from+1, expectedSort).Return(containers, int64(1), nil)
	s.logger.EXPECT().Info("containers exported successfully").Times(1)

	result, err := s.containerService.Export(s.ctx, filter, from, to, inputSort)
	s.NoError(err)
	s.NotNil(result)
	s.True(len(result) > 0)
}

func (s *ContainerServiceSuite) TestExportError() {
	filter := dto.ContainerFilter{}
	sort := dto.ContainerSort{Field: "container_id", Sort: "asc"}
	from, to := 1, 5

	s.mockRepo.EXPECT().View(filter, from, to-from+1, sort).Return(nil, int64(0), errors.New("fetch error"))

	_, err := s.containerService.Export(s.ctx, filter, from, to, sort)
	s.ErrorContains(err, "fetch error")
}

func (s *ContainerServiceSuite) TestExportWithEmptyData() {
	filter := dto.ContainerFilter{}
	sort := dto.ContainerSort{Field: "container_id", Sort: "asc"}
	from, to := 1, 10
	containers := []*entities.Container{}

	s.mockRepo.EXPECT().View(filter, from, to-from+1, sort).Return(containers, int64(0), nil)
	s.logger.EXPECT().Info(gomock.Any(), gomock.Any()).Times(1)

	result, err := s.containerService.Export(s.ctx, filter, from, to, sort)
	s.NoError(err)
	s.True(len(result) > 0)
}

func (s *ContainerServiceSuite) TestImportWithValidExcelFile() {
	s.logger.EXPECT().Error("failed to import containers", gomock.Any()).Times(1)

	// Create a minimal valid Excel-like structure
	data := []byte("PK\x03\x04\x14\x00\x00\x00\x08\x00") // ZIP header for Excel
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

	resp, err := s.containerService.Import(s.ctx, fakeFile)
	s.Error(err)
	s.Nil(resp)
}

func (s *ContainerServiceSuite) TestImportWithInvalidRows() {
	s.logger.EXPECT().Warn("skipping invalid row", gomock.Any()).Times(1)
	s.logger.EXPECT().Info("containers imported successfully").Times(1)

	f := excelize.NewFile()
	sheetName := "Sheet1"
	f.SetSheetName("Sheet1", sheetName)
	f.SetCellValue(sheetName, "A1", "Container ID")
	f.SetCellValue(sheetName, "B1", "Container Name")
	f.SetCellValue(sheetName, "C1", "Status")
	f.SetCellValue(sheetName, "D1", "IPv4")

	f.SetCellValue(sheetName, "A2", "test")

	var buf bytes.Buffer
	err := f.Write(&buf)
	s.Require().NoError(err)

	reader := bytes.NewReader(buf.Bytes())

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

	resp, err := s.containerService.Import(s.ctx, fakeFile)
	s.NoError(err)
	s.NotNil(resp)
	s.Equal(0, resp.SuccessCount)
	s.Equal(0, resp.FailedCount)
}

func (s *ContainerServiceSuite) TestImport() {
	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	f.SetCellValue(sheet, "A1", "Container ID")
	f.SetCellValue(sheet, "B1", "Container Name")
	f.SetCellValue(sheet, "C1", "Status")
	f.SetCellValue(sheet, "D1", "IPv4")

	f.SetCellValue(sheet, "A2", "test")
	f.SetCellValue(sheet, "B2", "container")
	f.SetCellValue(sheet, "C2", "ON")
	f.SetCellValue(sheet, "D2", "127.0.0.1")

	var buf bytes.Buffer
	err := f.Write(&buf)
	s.Require().NoError(err)

	reader := bytes.NewReader(buf.Bytes())
	file := struct {
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

	container := &entities.Container{
		ContainerId:   "test",
		ContainerName: "container",
		Status:        entities.ContainerOn,
		Ipv4:          "127.0.0.1",
	}
	s.mockRepo.EXPECT().Create("test", "container", entities.ContainerOn, "127.0.0.1").Return(container, nil)
	s.logger.EXPECT().Info("container created successfully", zap.String("containerId", container.ContainerId)).Times(1)
	s.logger.EXPECT().Info("containers imported successfully").Times(1)

	resp, err := s.containerService.Import(s.ctx, file)
	s.NoError(err)
	s.Equal(1, resp.SuccessCount)
	s.Equal(0, resp.FailedCount)
	s.Contains(resp.SuccessContainers, "test")
}

func (s *ContainerServiceSuite) TestImportCannotCreate() {
	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	f.SetCellValue(sheet, "A1", "Container ID")
	f.SetCellValue(sheet, "B1", "Container Name")
	f.SetCellValue(sheet, "C1", "Status")
	f.SetCellValue(sheet, "D1", "IPv4")

	f.SetCellValue(sheet, "A2", "test")
	f.SetCellValue(sheet, "B2", "container")
	f.SetCellValue(sheet, "C2", "ON")
	f.SetCellValue(sheet, "D2", "127.0.0.1")

	var buf bytes.Buffer
	err := f.Write(&buf)
	s.Require().NoError(err)

	reader := bytes.NewReader(buf.Bytes())
	file := struct {
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
	s.mockRepo.EXPECT().Create("test", "container", entities.ContainerOn, "127.0.0.1").Return(nil, errors.New("create error"))
	s.logger.EXPECT().Error("failed to create container", gomock.Any()).Times(1)
	s.logger.EXPECT().Info("containers imported successfully").Times(1)

	resp, err := s.containerService.Import(s.ctx, file)
	s.NoError(err)
	s.Equal(0, resp.SuccessCount)
	s.Equal(1, resp.FailedCount)
	s.Contains(resp.FailedContainers, "test")
}

func (s *ContainerServiceSuite) TestImportInvalidContainerField() {
	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	f.SetCellValue(sheet, "A1", "Container ID")
	f.SetCellValue(sheet, "B1", "Container Name")
	f.SetCellValue(sheet, "C1", "Status")
	f.SetCellValue(sheet, "D1", "IPv4")

	f.SetCellValue(sheet, "A2", "")
	f.SetCellValue(sheet, "B2", "container")
	f.SetCellValue(sheet, "C2", "ON")
	f.SetCellValue(sheet, "D2", "127.0.0.1")

	var buf bytes.Buffer
	err := f.Write(&buf)
	s.Require().NoError(err)

	reader := bytes.NewReader(buf.Bytes())
	file := struct {
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
	s.logger.EXPECT().Info("containers imported successfully").Times(1)

	resp, err := s.containerService.Import(s.ctx, file)
	s.NoError(err)
	s.Equal(0, resp.SuccessCount)
	s.Equal(1, resp.FailedCount)
	s.Contains(resp.FailedContainers, "")
}
