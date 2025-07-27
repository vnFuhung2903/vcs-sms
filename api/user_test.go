package api

import (
	"bytes"
	"encoding/json"
	"errors"
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

type UserHandlerSuite struct {
	suite.Suite
	ctrl              *gomock.Controller
	mockUserService   *services.MockIUserService
	mockJWTMiddleware *middlewares.MockIJWTMiddleware
	handler           *UserHandler
	router            *gin.Engine
}

func (s *UserHandlerSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockUserService = services.NewMockIUserService(s.ctrl)
	s.mockJWTMiddleware = middlewares.NewMockIJWTMiddleware(s.ctrl)

	s.mockJWTMiddleware.EXPECT().
		RequireScope(gomock.Any()).
		Return(func(c *gin.Context) {
			c.Set("userId", "test-user-id")
			c.Next()
		}).
		AnyTimes()

	s.handler = NewUserHandler(s.mockUserService, s.mockJWTMiddleware)

	gin.SetMode(gin.TestMode)
	s.router = gin.New()
	s.handler.SetupRoutes(s.router)
}

func (s *UserHandlerSuite) TearDownTest() {
	s.ctrl.Finish()
}

func TestUserHandlerSuite(t *testing.T) {
	suite.Run(t, new(UserHandlerSuite))
}

func (s *UserHandlerSuite) TestUpdateRole() {
	s.mockUserService.EXPECT().
		UpdateRole(gomock.Any(), "test-user-id", entities.Manager).
		Return(nil)

	reqBody := dto.UpdateRoleRequest{
		UserId: "test-user-id",
		Role:   entities.Manager,
	}
	jsonData, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("PUT", "/users/update/role", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusOK, w.Code)

	var response dto.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal("ROLE_UPDATED", response.Code)
}

func (s *UserHandlerSuite) TestUpdateRoleInvalidRequestBody() {
	req := httptest.NewRequest("PUT", "/users/update/role", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusBadRequest, w.Code)

	var response dto.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.NotEmpty(response.Error)
}

func (s *UserHandlerSuite) TestUpdateRoleServiceError() {
	s.mockUserService.EXPECT().
		UpdateRole(gomock.Any(), "test-user-id", entities.Manager).
		Return(errors.New("service error"))

	reqBody := dto.UpdateRoleRequest{
		UserId: "test-user-id",
		Role:   entities.Manager,
	}
	jsonData, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("PUT", "/users/update/role", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusInternalServerError, w.Code)

	var response dto.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal("service error", response.Error)
}

func (s *UserHandlerSuite) TestUpdateScope() {
	scopes := []string{"container:create", "container:update"}
	s.mockUserService.EXPECT().
		UpdateScope(gomock.Any(), "test-user-id", scopes, true).
		Return(nil)

	reqBody := dto.UpdateScopeRequest{
		UserId:  "test-user-id",
		IsAdded: true,
		Scopes:  scopes,
	}
	jsonData, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("PUT", "/users/update/scope", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusOK, w.Code)

	var response dto.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal("SCOPE_UPDATED", response.Code)
}

func (s *UserHandlerSuite) TestUpdateScopeInvalidRequestBody() {
	req := httptest.NewRequest("PUT", "/users/update/scope", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusBadRequest, w.Code)

	var response dto.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.NotEmpty(response.Error)
}

func (s *UserHandlerSuite) TestUpdateScopeServiceError() {
	scopes := []string{"container:create"}
	s.mockUserService.EXPECT().
		UpdateScope(gomock.Any(), "test-user-id", scopes, false).
		Return(errors.New("service error"))

	reqBody := dto.UpdateScopeRequest{
		UserId:  "test-user-id",
		IsAdded: false,
		Scopes:  scopes,
	}
	jsonData, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("PUT", "/users/update/scope", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusInternalServerError, w.Code)

	var response dto.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal("service error", response.Error)
}

func (s *UserHandlerSuite) TestDelete() {
	s.mockUserService.EXPECT().
		Delete(gomock.Any(), "test-user-id").
		Return(nil)

	reqBody := dto.DeleteRequest{
		UserId: "test-user-id",
	}
	jsonData, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("DELETE", "/users/delete", bytes.NewBuffer(jsonData))
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusOK, w.Code)

	var response dto.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal("USER_DELETED", response.Code)
}

func (s *UserHandlerSuite) TestDeleteInvalidRequestBody() {
	req := httptest.NewRequest("DELETE", "/users/delete", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusBadRequest, w.Code)

	var response dto.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.NotEmpty(response.Error)
}

func (s *UserHandlerSuite) TestDeleteServiceError() {
	s.mockUserService.EXPECT().
		Delete(gomock.Any(), "test-user-id").
		Return(errors.New("service error"))

	reqBody := dto.DeleteRequest{
		UserId: "test-user-id",
	}
	jsonData, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("DELETE", "/users/delete", bytes.NewBuffer(jsonData))
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusInternalServerError, w.Code)

	var response dto.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal("service error", response.Error)
}
