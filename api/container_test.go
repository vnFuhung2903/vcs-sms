package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"github.com/vnFuhung2903/vcs-sms/dto"
	"github.com/vnFuhung2903/vcs-sms/entities"
	"github.com/vnFuhung2903/vcs-sms/mocks/middlewares"
	"github.com/vnFuhung2903/vcs-sms/mocks/services"
)

type ContainerHandlerSuite struct {
	suite.Suite
	ctrl                 *gomock.Controller
	mockContainerService *services.MockIContainerService
	mockJWTMiddleware    *middlewares.MockIJWTMiddleware
	handler              *ContainerHandler
	router               *gin.Engine
	ctx                  context.Context
}

func (s *ContainerHandlerSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockContainerService = services.NewMockIContainerService(s.ctrl)
	s.mockJWTMiddleware = middlewares.NewMockIJWTMiddleware(s.ctrl)
	s.ctx = context.Background()

	s.mockJWTMiddleware.EXPECT().
		RequireScope(gomock.Any()).
		Return(func(c *gin.Context) {
			c.Next()
		}).
		AnyTimes()

	s.handler = NewContainerHandler(s.mockContainerService, s.mockJWTMiddleware)

	gin.SetMode(gin.TestMode)
	s.router = gin.New()
	s.handler.SetupRoutes(s.router)
}
func (s *ContainerHandlerSuite) TearDownTest() {
	s.ctrl.Finish()
}

func TestContainerHandlerSuite(t *testing.T) {
	suite.Run(t, new(ContainerHandlerSuite))
}

func (s *ContainerHandlerSuite) TestCreate() {
	container := &entities.Container{
		ContainerId:   "1",
		ContainerName: "test-container",
		Ipv4:          "127.0.0.1",
		Status:        "running",
	}

	s.mockContainerService.EXPECT().
		Create(gomock.Any(), "test-container", "nginx").
		Return(container, nil)

	reqBody := dto.CreateRequest{
		ContainerName: "test-container",
		ImageName:     "nginx",
	}
	jsonData, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/containers/create", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
}

func (s *ContainerHandlerSuite) TestCreateInvalidRequestBody() {
	req := httptest.NewRequest("POST", "/containers/create", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code)

	var response dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.NotEmpty(response.Error)
}

func (s *ContainerHandlerSuite) TestCreateServiceError() {
	s.mockContainerService.EXPECT().
		Create(gomock.Any(), "test-container", "nginx").
		Return((*entities.Container)(nil), errors.New("service error"))

	reqBody := dto.CreateRequest{
		ContainerName: "test-container",
		ImageName:     "nginx",
	}
	jsonData, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/containers/create", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusInternalServerError, w.Code)

	var response dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal("service error", response.Error)
}

func (s *ContainerHandlerSuite) TestView() {
	containers := []*entities.Container{
		{ContainerId: "1", ContainerName: "container1", Status: "running"},
		{ContainerId: "2", ContainerName: "container2", Status: "stopped"},
	}

	s.mockContainerService.EXPECT().
		View(gomock.Any(), gomock.Any(), 1, 10, gomock.Any()).
		Return(containers, int64(2), nil)

	req := httptest.NewRequest("GET", "/containers/view?from=1&to=10&field=container_id&order=desc", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var response dto.ViewResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal(2, len(response.Data))
	s.Equal(int64(2), response.Total)
}

func (s *ContainerHandlerSuite) TestViewInvalidFromParameter() {
	req := httptest.NewRequest("GET", "/containers/view?from=invalid-number", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code)

	var response dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Contains(response.Error, "invalid syntax")
}

func (s *ContainerHandlerSuite) TestViewInvalidToParameter() {
	req := httptest.NewRequest("GET", "/containers/view?from=1&to=invalid-number", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code)

	var response dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Contains(response.Error, "invalid syntax")
}

func (s *ContainerHandlerSuite) TestViewInvalidFilterParameter() {
	req := httptest.NewRequest("GET", "/containers/view?status=invalid-status", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code)

	var response dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Contains(response.Error, "Field validation for 'Status' failed on the 'oneof' tag")
}

func (s *ContainerHandlerSuite) TestViewInvalidSortParameter() {
	req := httptest.NewRequest("GET", "/containers/view?order=invalid-order", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code)

	var response dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Contains(response.Error, "Field validation for 'Order' failed on the 'oneof' tag")
}

func (s *ContainerHandlerSuite) TestViewServiceError() {
	s.mockContainerService.EXPECT().
		View(gomock.Any(), gomock.Any(), 1, 10, gomock.Any()).
		Return([]*entities.Container{}, int64(0), errors.New("database error"))

	req := httptest.NewRequest("GET", "/containers/view?from=1&to=10&field=container_id&order=desc", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusInternalServerError, w.Code)

	var response dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal("database error", response.Error)
}

func (s *ContainerHandlerSuite) TestUpdate() {
	s.mockContainerService.EXPECT().
		Update(gomock.Any(), "container-id", gomock.Any()).
		Return(nil)

	updateData := dto.ContainerUpdate{
		Status: entities.ContainerOff,
	}
	jsonData, _ := json.Marshal(updateData)

	req := httptest.NewRequest("PUT", "/containers/update/container-id", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
}

func (s *ContainerHandlerSuite) TestUpdateInvalidRequestBody() {
	req := httptest.NewRequest("PUT", "/containers/update/container-id", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code)

	var response dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.NotEmpty(response.Error)
}

func (s *ContainerHandlerSuite) TestUpdateServiceError() {
	s.mockContainerService.EXPECT().
		Update(gomock.Any(), "container-id", gomock.Any()).
		Return(errors.New("update failed"))

	updateData := dto.ContainerUpdate{
		Status: entities.ContainerOff,
	}
	jsonData, _ := json.Marshal(updateData)

	req := httptest.NewRequest("PUT", "/containers/update/container-id", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusInternalServerError, w.Code)

	var response dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal("update failed", response.Error)
}

func (s *ContainerHandlerSuite) TestDelete() {
	s.mockContainerService.EXPECT().
		Delete(gomock.Any(), "container-id").
		Return(nil)

	req := httptest.NewRequest("DELETE", "/containers/delete/container-id", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
}

func (s *ContainerHandlerSuite) TestDeleteServiceError() {
	s.mockContainerService.EXPECT().
		Delete(gomock.Any(), "container-id").
		Return(errors.New("delete failed"))

	req := httptest.NewRequest("DELETE", "/containers/delete/container-id", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusInternalServerError, w.Code)

	var response dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal("delete failed", response.Error)
}

func (s *ContainerHandlerSuite) TestExport() {
	csvData := []byte("id,name,status\n1,container1,running")

	s.mockContainerService.EXPECT().
		Export(gomock.Any(), gomock.Any(), 1, -1, gomock.Any()).
		Return(csvData, nil)

	req := httptest.NewRequest("GET", "/containers/export?field=container_id&order=desc", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
	s.Equal("application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", w.Header().Get("Content-Type"))
	s.Equal("attachment; filename=\"containers.xlsx\"", w.Header().Get("Content-Disposition"))
	s.Equal(csvData, w.Body.Bytes())
}

func (s *ContainerHandlerSuite) TestExportInvalidFromParameter() {
	req := httptest.NewRequest("GET", "/containers/export?from=invalid", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code)

	var response dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Contains(response.Error, "invalid syntax")
}

func (s *ContainerHandlerSuite) TestExportInvalidToParameter() {
	req := httptest.NewRequest("GET", "/containers/export?from=1&to=invalid-number", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code)

	var response dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Contains(response.Error, "invalid syntax")
}

func (s *ContainerHandlerSuite) TestExportServiceError() {
	s.mockContainerService.EXPECT().
		Export(gomock.Any(), gomock.Any(), 1, -1, gomock.Any()).
		Return([]byte{}, errors.New("export failed"))

	req := httptest.NewRequest("GET", "/containers/export?field=container_id&order=desc", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusInternalServerError, w.Code)

	var response dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal("export failed", response.Error)
}

func (s *ContainerHandlerSuite) TestExportInvalidFilterParameter() {
	req := httptest.NewRequest("GET", "/containers/export?status=invalid-status", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code)

	var response dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Contains(response.Error, "Field validation for 'Status' failed on the 'oneof' tag")
}

func (s *ContainerHandlerSuite) TestExportInvalidSortParameter() {
	req := httptest.NewRequest("GET", "/containers/export?order=invalid-order", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code)

	var response dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Contains(response.Error, "Field validation for 'Order' failed on the 'oneof' tag")
}

func (s *ContainerHandlerSuite) TestImport() {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.csv")
	part.Write([]byte("id,name,status\n1,test,running"))
	writer.Close()

	response := &dto.ImportResponse{
		SuccessCount:      1,
		SuccessContainers: []string{"test"},
		FailedCount:       0,
		FailedContainers:  []string{},
	}

	s.mockContainerService.EXPECT().
		Import(gomock.Any(), gomock.Any()).
		Return(response, nil)

	req := httptest.NewRequest("POST", "/containers/import", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal(1, response.SuccessCount)
	s.Equal(0, response.FailedCount)
}

func (s *ContainerHandlerSuite) TestImportMissingFile() {
	req := httptest.NewRequest("POST", "/containers/import", nil)
	req.Header.Set("Content-Type", "multipart/form-data")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code)

	var response dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Contains(response.Error, "no multipart boundary param in Content-Type")
}

func (s *ContainerHandlerSuite) TestImportServiceError() {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.csv")
	part.Write([]byte("id,name,status\n1,test,running"))
	writer.Close()

	s.mockContainerService.EXPECT().
		Import(gomock.Any(), gomock.Any()).
		Return((*dto.ImportResponse)(nil), errors.New("import failed"))

	req := httptest.NewRequest("POST", "/containers/import", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusInternalServerError, w.Code)

	var response dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal("import failed", response.Error)
}
