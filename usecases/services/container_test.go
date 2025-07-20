package services

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"

	"github.com/vnFuhung2903/vcs-sms/dto"
	"github.com/vnFuhung2903/vcs-sms/entities"
	"github.com/vnFuhung2903/vcs-sms/mocks/docker"
	"github.com/vnFuhung2903/vcs-sms/mocks/logger"
	"github.com/vnFuhung2903/vcs-sms/mocks/repositories"
)

type ContainerServiceSuite struct {
	suite.Suite
	ctrl             *gomock.Controller
	containerService IContainerService
	mockRepo         *repositories.MockIContainerRepository
	dockerClient     *docker.MockIDockerClient
	logger           *logger.MockILogger
	ctx              context.Context
}

func (s *ContainerServiceSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockRepo = repositories.NewMockIContainerRepository(s.ctrl)
	s.dockerClient = docker.NewMockIDockerClient(s.ctrl)
	s.logger = logger.NewMockILogger(s.ctrl)
	s.containerService = NewContainerService(s.mockRepo, s.dockerClient, s.logger)
	s.ctx = context.Background()
}

func (s *ContainerServiceSuite) TearDownTest() {
	s.ctrl.Finish()
}

func TestContainerServiceSuite(t *testing.T) {
	suite.Run(t, new(ContainerServiceSuite))
}

func (s *ContainerServiceSuite) TestCreate() {
	containerResp := &container.CreateResponse{ID: "test-id"}

	s.dockerClient.EXPECT().Create(s.ctx, "container", "testcontainers/ryuk:0.12.0").Return(containerResp, nil)
	s.dockerClient.EXPECT().Start(s.ctx, "test-id").Return(nil)
	s.dockerClient.EXPECT().GetStatus(s.ctx, "test-id").Return(entities.ContainerOn)
	s.dockerClient.EXPECT().GetIpv4(s.ctx, "test-id").Return("127.0.0.1")
	s.mockRepo.EXPECT().Create("test-id", "container", entities.ContainerOn, "127.0.0.1").Return(&entities.Container{
		ContainerId:   "test-id",
		ContainerName: "container",
		Status:        entities.ContainerOn,
		Ipv4:          "127.0.0.1",
	}, nil)
	s.logger.EXPECT().Info("container created successfully", zap.String("containerId", "test-id")).Times(1)

	result, err := s.containerService.Create(s.ctx, "container", "testcontainers/ryuk:0.12.0")
	s.NoError(err)
	s.Equal("container", result.ContainerName)
	s.Equal("test-id", result.ContainerId)
}

func (s *ContainerServiceSuite) TestCreateDockerCreateError() {
	s.dockerClient.EXPECT().Create(s.ctx, "container", "testcontainers/ryuk:0.12.0").Return(nil, errors.New("docker create error"))
	s.logger.EXPECT().Error("failed to create container", gomock.Any()).Times(1)

	result, err := s.containerService.Create(s.ctx, "container", "testcontainers/ryuk:0.12.0")
	s.ErrorContains(err, "docker create error")
	s.Nil(result)
}

func (s *ContainerServiceSuite) TestCreateDockerStartError() {
	containerResp := &container.CreateResponse{ID: "test-id"}

	s.dockerClient.EXPECT().Create(s.ctx, "container", "testcontainers/ryuk:0.12.0").Return(containerResp, nil)
	s.dockerClient.EXPECT().Start(s.ctx, "test-id").Return(errors.New("docker start error"))
	s.dockerClient.EXPECT().GetStatus(s.ctx, "test-id").Return(entities.ContainerOff)
	s.dockerClient.EXPECT().GetIpv4(s.ctx, "test-id").Return("")
	s.mockRepo.EXPECT().Create("test-id", "container", entities.ContainerOff, "").Return(&entities.Container{
		ContainerId:   "test-id",
		ContainerName: "container",
		Status:        entities.ContainerOff,
		Ipv4:          "",
	}, nil)
	s.logger.EXPECT().Error("failed to start container", zap.Error(errors.New("docker start error"))).Times(1)
	s.logger.EXPECT().Info("container created successfully", zap.String("containerId", "test-id")).Times(1)

	result, err := s.containerService.Create(s.ctx, "container", "testcontainers/ryuk:0.12.0")
	s.NoError(err)
	s.Equal("container", result.ContainerName)
	s.Equal("test-id", result.ContainerId)
}

func (s *ContainerServiceSuite) TestCreateRepoError() {
	containerResp := &container.CreateResponse{ID: "test-id"}

	s.dockerClient.EXPECT().Create(s.ctx, "container", "testcontainers/ryuk:0.12.0").Return(containerResp, nil)
	s.dockerClient.EXPECT().Start(s.ctx, "test-id").Return(nil)
	s.dockerClient.EXPECT().GetStatus(s.ctx, "test-id").Return(entities.ContainerOn)
	s.dockerClient.EXPECT().GetIpv4(s.ctx, "test-id").Return("127.0.0.1")
	s.dockerClient.EXPECT().Delete(s.ctx, "test-id").Return(nil)
	s.mockRepo.EXPECT().Create("test-id", "container", entities.ContainerOn, "127.0.0.1").Return(nil, errors.New("db error"))
	s.logger.EXPECT().Error("failed to create container", gomock.Any()).Times(1)

	result, err := s.containerService.Create(s.ctx, "container", "testcontainers/ryuk:0.12.0")
	s.ErrorContains(err, "db error")
	s.Nil(result)
}

func (s *ContainerServiceSuite) TestCreateRepoAndDockerDeleteError() {
	containerResp := &container.CreateResponse{ID: "test-id"}

	s.dockerClient.EXPECT().Create(s.ctx, "container", "testcontainers/ryuk:0.12.0").Return(containerResp, nil)
	s.dockerClient.EXPECT().Start(s.ctx, "test-id").Return(nil)
	s.dockerClient.EXPECT().GetStatus(s.ctx, "test-id").Return(entities.ContainerOn)
	s.dockerClient.EXPECT().GetIpv4(s.ctx, "test-id").Return("127.0.0.1")
	s.dockerClient.EXPECT().Delete(s.ctx, "test-id").Return(errors.New("docker delete error"))
	s.mockRepo.EXPECT().Create("test-id", "container", entities.ContainerOn, "127.0.0.1").Return(nil, errors.New("db error"))
	s.logger.EXPECT().Error("failed to create container", gomock.Any()).Times(1)
	s.logger.EXPECT().Error("failed to delete container", gomock.Any()).Times(1)

	result, err := s.containerService.Create(s.ctx, "container", "testcontainers/ryuk:0.12.0")
	s.ErrorContains(err, "docker delete error")
	s.Nil(result)
}

func (s *ContainerServiceSuite) TestView() {
	filter := dto.ContainerFilter{}
	sort := dto.ContainerSort{Field: "container_id", Order: "asc"}
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

func (s *ContainerServiceSuite) TestUpdateOn() {
	updateData := dto.ContainerUpdate{Status: "ON"}

	s.dockerClient.EXPECT().Start(s.ctx, "test-id").Return(nil)
	s.mockRepo.EXPECT().Update("test-id", updateData).Return(nil)
	s.logger.EXPECT().Info("container updated successfully", gomock.Any()).Times(1)

	err := s.containerService.Update(s.ctx, "test-id", updateData)
	s.NoError(err)
}

func (s *ContainerServiceSuite) TestUpdateOff() {
	updateData := dto.ContainerUpdate{Status: "OFF"}

	s.dockerClient.EXPECT().Stop(s.ctx, "test-id").Return(nil)
	s.mockRepo.EXPECT().Update("test-id", updateData).Return(nil)
	s.logger.EXPECT().Info("container updated successfully", gomock.Any()).Times(1)

	err := s.containerService.Update(s.ctx, "test-id", updateData)
	s.NoError(err)
}

func (s *ContainerServiceSuite) TestUpdateInvalidStatus() {
	updateData := dto.ContainerUpdate{Status: "INVALID"}

	err := s.containerService.Update(s.ctx, "test-id", updateData)
	s.ErrorContains(err, "invalid status")
}

func (s *ContainerServiceSuite) TestUpdateDockerStartError() {
	updateData := dto.ContainerUpdate{Status: "ON"}

	s.dockerClient.EXPECT().Start(s.ctx, "test-id").Return(errors.New("docker start failed"))
	s.logger.EXPECT().Error("failed to start container", gomock.Any()).Times(1)

	err := s.containerService.Update(s.ctx, "test-id", updateData)
	s.ErrorContains(err, "docker start failed")
}

func (s *ContainerServiceSuite) TestUpdateDockerStopError() {
	updateData := dto.ContainerUpdate{Status: "OFF"}

	s.dockerClient.EXPECT().Stop(s.ctx, "test-id").Return(errors.New("docker stop failed"))
	s.logger.EXPECT().Error("failed to stop container", gomock.Any()).Times(1)

	err := s.containerService.Update(s.ctx, "test-id", updateData)
	s.ErrorContains(err, "docker stop failed")
}

func (s *ContainerServiceSuite) TestUpdateRepoError() {
	updateData := dto.ContainerUpdate{Status: "OFF"}

	s.dockerClient.EXPECT().Stop(s.ctx, "test-id").Return(nil)
	s.mockRepo.EXPECT().Update("test-id", updateData).Return(errors.New("update failed"))
	s.logger.EXPECT().Error("failed to update container", gomock.Any()).Times(1)

	err := s.containerService.Update(s.ctx, "test-id", updateData)
	s.ErrorContains(err, "update failed")
}

func (s *ContainerServiceSuite) TestDelete() {
	s.dockerClient.EXPECT().Stop(s.ctx, "test-id").Return(nil)
	s.dockerClient.EXPECT().Delete(s.ctx, "test-id").Return(nil)
	s.mockRepo.EXPECT().Delete("test-id").Return(nil)
	s.logger.EXPECT().Info("container deleted successfully", zap.String("containerId", "test-id")).Times(1)

	err := s.containerService.Delete(s.ctx, "test-id")
	s.NoError(err)
}

func (s *ContainerServiceSuite) TestDeleteDockerStopError() {
	s.dockerClient.EXPECT().Stop(s.ctx, "test-id").Return(errors.New("stop failed"))
	s.logger.EXPECT().Error(gomock.Any(), gomock.Any()).Times(1)

	err := s.containerService.Delete(s.ctx, "test-id")
	s.ErrorContains(err, "stop failed")
}

func (s *ContainerServiceSuite) TestDeleteDockerDeleteError() {
	s.dockerClient.EXPECT().Stop(s.ctx, "test-id").Return(nil)
	s.dockerClient.EXPECT().Delete(s.ctx, "test-id").Return(errors.New("delete failed"))
	s.logger.EXPECT().Error(gomock.Any(), gomock.Any()).Times(1)

	err := s.containerService.Delete(s.ctx, "test-id")
	s.ErrorContains(err, "delete failed")
}

func (s *ContainerServiceSuite) TestDeleteRepoError() {
	s.dockerClient.EXPECT().Stop(s.ctx, "test-id").Return(nil)
	s.dockerClient.EXPECT().Delete(s.ctx, "test-id").Return(nil)
	s.mockRepo.EXPECT().Delete("test-id").Return(errors.New("delete failed"))
	s.logger.EXPECT().Error("failed to delete container", gomock.Any()).Times(1)

	err := s.containerService.Delete(s.ctx, "test-id")
	s.ErrorContains(err, "delete failed")
}

func (s *ContainerServiceSuite) TestImport() {
	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	f.SetCellValue(sheet, "A1", "Container Name")
	f.SetCellValue(sheet, "B1", "Image Name")

	f.SetCellValue(sheet, "A2", "test-name")
	f.SetCellValue(sheet, "B2", "nginx")

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

	containerResp := &container.CreateResponse{ID: "test-id"}
	containerEntity := &entities.Container{
		ContainerId:   "test-id",
		ContainerName: "test-name",
		Status:        entities.ContainerOn,
		Ipv4:          "127.0.0.1",
	}

	s.dockerClient.EXPECT().Create(s.ctx, "test-name", "nginx").Return(containerResp, nil)
	s.dockerClient.EXPECT().Start(s.ctx, "test-id").Return(nil)
	s.dockerClient.EXPECT().GetStatus(s.ctx, "test-id").Return(entities.ContainerOn)
	s.dockerClient.EXPECT().GetIpv4(s.ctx, "test-id").Return("127.0.0.1")
	s.mockRepo.EXPECT().Create("test-id", "test-name", entities.ContainerOn, "127.0.0.1").Return(containerEntity, nil)
	s.logger.EXPECT().Info("container created successfully", zap.String("containerId", "test-id")).Times(1)
	s.logger.EXPECT().Info("containers imported successfully").Times(1)

	resp, err := s.containerService.Import(s.ctx, file)
	s.NoError(err)
	s.Equal(1, resp.SuccessCount)
	s.Equal(0, resp.FailedCount)
	s.Contains(resp.SuccessContainers, "test-name")
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

func (s *ContainerServiceSuite) TestImportWithMissingHeaderRows() {
	s.logger.EXPECT().Error("failed to import containers", gomock.Any()).Times(1)

	f := excelize.NewFile()
	sheetName := "Sheet1"
	f.SetSheetName("Sheet1", sheetName)
	f.SetCellValue(sheetName, "A1", "Container Id")

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
	s.Error(err)
	s.Nil(resp)
}

func (s *ContainerServiceSuite) TestImportWithInvalidHeaderRows() {
	s.logger.EXPECT().Error("failed to import containers", gomock.Any()).Times(1)

	f := excelize.NewFile()
	sheetName := "Sheet1"
	f.SetSheetName("Sheet1", sheetName)
	f.SetCellValue(sheetName, "A1", "Container Id")
	f.SetCellValue(sheetName, "B1", "Image Name")

	f.SetCellValue(sheetName, "A2", "test-id")

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
	s.Error(err)
	s.Nil(resp)
}

func (s *ContainerServiceSuite) TestImportWithInvalidRows() {
	s.logger.EXPECT().Warn("skipping invalid row", gomock.Any()).Times(1)
	s.logger.EXPECT().Info("containers imported successfully").Times(1)

	f := excelize.NewFile()
	sheetName := "Sheet1"
	f.SetSheetName("Sheet1", sheetName)
	f.SetCellValue(sheetName, "A1", "Container Name")
	f.SetCellValue(sheetName, "B1", "Image Name")

	f.SetCellValue(sheetName, "A2", "test-name")

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

func (s *ContainerServiceSuite) TestImportCreateError() {
	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	f.SetCellValue(sheet, "A1", "Container Name")
	f.SetCellValue(sheet, "B1", "Image Name")

	f.SetCellValue(sheet, "A2", "test-name")
	f.SetCellValue(sheet, "B2", "nginx")

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

	s.dockerClient.EXPECT().Create(s.ctx, "test-name", "nginx").Return(nil, errors.New("create error"))
	s.logger.EXPECT().Error("failed to create container", gomock.Any()).Times(1)
	s.logger.EXPECT().Info("containers imported successfully").Times(1)

	resp, err := s.containerService.Import(s.ctx, file)
	s.NoError(err)
	s.Equal(0, resp.SuccessCount)
	s.Equal(1, resp.FailedCount)
	s.Contains(resp.FailedContainers, "test-name")
}

func (s *ContainerServiceSuite) TestImportInvalidContainerField() {
	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	f.SetCellValue(sheet, "A1", "Container Name")
	f.SetCellValue(sheet, "B1", "Image Name")

	f.SetCellValue(sheet, "A2", "")
	f.SetCellValue(sheet, "B2", "nginx")

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

func (s *ContainerServiceSuite) TestExport() {
	filter := dto.ContainerFilter{}
	sort := dto.ContainerSort{Field: "container_id", Order: "asc"}
	from, to := 1, 10
	containers := []*entities.Container{
		{
			ContainerId:   "abc",
			ContainerName: "test-id",
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

func (s *ContainerServiceSuite) TestExportError() {
	filter := dto.ContainerFilter{}
	sort := dto.ContainerSort{Field: "container_id", Order: "asc"}
	from, to := 1, 5

	s.mockRepo.EXPECT().View(filter, from, to-from+1, sort).Return(nil, int64(0), errors.New("fetch error"))

	_, err := s.containerService.Export(s.ctx, filter, from, to, sort)
	s.ErrorContains(err, "fetch error")
}
