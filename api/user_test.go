package api

import (
	"bytes"
	"context"
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
	ctx               context.Context
}

func (s *UserHandlerSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockUserService = services.NewMockIUserService(s.ctrl)
	s.mockJWTMiddleware = middlewares.NewMockIJWTMiddleware(s.ctrl)
	s.ctx = context.Background()

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

func (s *UserHandlerSuite) TestRegister() {
	user := &entities.User{
		ID:       "1",
		Username: "testuser",
		Email:    "test@example.com",
		Role:     entities.Developer,
		Scopes:   5,
	}

	s.mockUserService.EXPECT().
		Register("testuser", "password123", "test@example.com", entities.Developer, gomock.Any()).
		Return(user, nil)

	s.mockJWTMiddleware.EXPECT().
		GenerateJWT(gomock.Any(), "1", "testuser", gomock.Any()).
		Return(nil)

	reqBody := dto.RegisterRequest{
		Username: "testuser",
		Password: "password123",
		Email:    "test@example.com",
		Role:     entities.Developer,
	}
	jsonData, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/users/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
}

func (s *UserHandlerSuite) TestRegisterInvalidRequestBody() {
	req := httptest.NewRequest("POST", "/users/register", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code)

	var response dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.NotEmpty(response.Error)
}

func (s *UserHandlerSuite) TestRegisterServiceError() {
	s.mockUserService.EXPECT().
		Register("testuser", "password123", "test@example.com", entities.Developer, gomock.Any()).
		Return((*entities.User)(nil), errors.New("registration failed"))

	reqBody := dto.RegisterRequest{
		Username: "testuser",
		Password: "password123",
		Email:    "test@example.com",
		Role:     entities.Developer,
	}
	jsonData, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/users/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusInternalServerError, w.Code)

	var response dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal("registration failed", response.Error)
}

func (s *UserHandlerSuite) TestRegisterJWTGenerationError() {
	user := &entities.User{
		ID:       "1",
		Username: "testuser",
		Hash:     "hashedpassword",
		Email:    "test@example.com",
		Role:     entities.Developer,
		Scopes:   int64(5),
	}

	s.mockUserService.EXPECT().
		Register("testuser", "password123", "test@example.com", entities.Developer, gomock.Any()).
		Return(user, nil)

	s.mockJWTMiddleware.EXPECT().
		GenerateJWT(gomock.Any(), "1", "testuser", gomock.Any()).
		Return(errors.New("JWT generation failed"))

	reqBody := dto.RegisterRequest{
		Username: "testuser",
		Password: "password123",
		Email:    "test@example.com",
		Role:     entities.Developer,
	}
	jsonData, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/users/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusInternalServerError, w.Code)

	var response dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal("JWT generation failed", response.Error)
}

func (s *UserHandlerSuite) TestLogin() {
	user := &entities.User{
		ID:       "1",
		Username: "testuser",
		Hash:     "hashedpassword",
		Email:    "test@example.com",
		Role:     entities.Developer,
		Scopes:   int64(5),
	}

	s.mockUserService.EXPECT().
		Login("testuser", "password123").
		Return(user, nil)

	s.mockJWTMiddleware.EXPECT().
		GenerateJWT(gomock.Any(), "1", "testuser", gomock.Any()).
		Return(nil)

	reqBody := dto.LoginRequest{
		Username: "testuser",
		Password: "password123",
	}
	jsonData, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/users/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
}

func (s *UserHandlerSuite) TestLoginInvalidRequestBody() {
	req := httptest.NewRequest("POST", "/users/login", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code)

	var response dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.NotEmpty(response.Error)
}

func (s *UserHandlerSuite) TestLoginInvalidCredentials() {
	s.mockUserService.EXPECT().
		Login("testuser", "wrongpassword").
		Return((*entities.User)(nil), errors.New("invalid credentials"))

	reqBody := dto.LoginRequest{
		Username: "testuser",
		Password: "wrongpassword",
	}
	jsonData, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/users/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusUnauthorized, w.Code)

	var response dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal("invalid credentials", response.Error)
}

func (s *UserHandlerSuite) TestLoginJWTGenerationError() {
	user := &entities.User{
		ID:       "1",
		Username: "testuser",
		Scopes:   5,
	}

	s.mockUserService.EXPECT().
		Login("testuser", "password123").
		Return(user, nil)

	s.mockJWTMiddleware.EXPECT().
		GenerateJWT(gomock.Any(), "1", "testuser", gomock.Any()).
		Return(errors.New("JWT generation failed"))

	reqBody := dto.LoginRequest{
		Username: "testuser",
		Password: "password123",
	}
	jsonData, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/users/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusInternalServerError, w.Code)

	var response dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal("JWT generation failed", response.Error)
}

func (s *UserHandlerSuite) TestUpdatePassword() {
	s.mockUserService.EXPECT().
		UpdatePassword("test-user-id", "newpassword").
		Return(nil)

	reqBody := dto.UpdatePasswordRequest{
		Password: "newpassword",
	}
	jsonData, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("PUT", "/users/update/password/test-user-id", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
}

func (s *UserHandlerSuite) TestUpdatePasswordInvalidRequestBody() {
	req := httptest.NewRequest("PUT", "/users/update/password/test-user-id", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code)

	var response dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.NotEmpty(response.Error)
}

func (s *UserHandlerSuite) TestUpdatePasswordServiceError() {
	s.mockUserService.EXPECT().
		UpdatePassword("test-user-id", "newpassword").
		Return(errors.New("password update failed"))

	reqBody := struct {
		Password string `json:"password"`
	}{
		Password: "newpassword",
	}
	jsonData, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("PUT", "/users/update/password/test-user-id", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusInternalServerError, w.Code)

	var response dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal("password update failed", response.Error)
}

func (s *UserHandlerSuite) TestUpdateRole() {
	s.mockUserService.EXPECT().
		UpdateRole("test-user-id", entities.Manager).
		Return(nil)

	reqBody := struct {
		Role string `json:"role"`
	}{
		Role: string(entities.Manager),
	}
	jsonData, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("PUT", "/users/update/role/test-user-id", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
}

func (s *UserHandlerSuite) TestUpdateRoleInvalidRequestBody() {
	req := httptest.NewRequest("PUT", "/users/update/role/test-user-id", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code)

	var response dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.NotEmpty(response.Error)
}

func (s *UserHandlerSuite) TestUpdateRoleServiceError() {
	s.mockUserService.EXPECT().
		UpdateRole("test-user-id", entities.Manager).
		Return(errors.New("role update failed"))

	reqBody := dto.UpdateRoleRequest{
		Role: entities.Manager,
	}
	jsonData, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("PUT", "/users/update/role/test-user-id", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusInternalServerError, w.Code)

	var response dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal("role update failed", response.Error)
}

func (s *UserHandlerSuite) TestUpdateScope() {
	scopes := []string{"container:create", "container:update"}
	s.mockUserService.EXPECT().
		UpdateScope("test-user-id", scopes, true).
		Return(nil)

	reqBody := dto.UpdateScopeRequest{
		IsAdded: true,
		Scopes:  scopes,
	}
	jsonData, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("PUT", "/users/update/scope/test-user-id", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
}

func (s *UserHandlerSuite) TestUpdateScopeInvalidRequestBody() {
	req := httptest.NewRequest("PUT", "/users/update/scope/test-user-id", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code)

	var response dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.NotEmpty(response.Error)
}

func (s *UserHandlerSuite) TestUpdateScopeServiceError() {
	scopes := []string{"container:create"}
	s.mockUserService.EXPECT().
		UpdateScope("test-user-id", scopes, false).
		Return(errors.New("scope update failed"))

	reqBody := dto.UpdateScopeRequest{
		IsAdded: false,
		Scopes:  scopes,
	}
	jsonData, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("PUT", "/users/update/scope/test-user-id", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusInternalServerError, w.Code)

	var response dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal("scope update failed", response.Error)
}

func (s *UserHandlerSuite) TestDelete() {
	s.mockUserService.EXPECT().
		Delete("test-user-id").
		Return(nil)

	req := httptest.NewRequest("DELETE", "/users/delete/test-user-id", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
}

func (s *UserHandlerSuite) TestDeleteServiceError() {
	s.mockUserService.EXPECT().
		Delete("test-user-id").
		Return(errors.New("delete failed"))

	req := httptest.NewRequest("DELETE", "/users/delete/test-user-id", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusInternalServerError, w.Code)

	var response dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal("delete failed", response.Error)
}
