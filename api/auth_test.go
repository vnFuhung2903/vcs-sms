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

type AuthHandlerSuite struct {
	suite.Suite
	ctrl              *gomock.Controller
	mockAuthService   *services.MockIAuthService
	mockJWTMiddleware *middlewares.MockIJWTMiddleware
	handler           *AuthHandler
	router            *gin.Engine
}

func (s *AuthHandlerSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockAuthService = services.NewMockIAuthService(s.ctrl)
	s.mockJWTMiddleware = middlewares.NewMockIJWTMiddleware(s.ctrl)

	s.mockJWTMiddleware.EXPECT().
		RequireScope(gomock.Any()).
		Return(func(c *gin.Context) {
			c.Set("userId", "test-user-id")
			c.Next()
		}).
		AnyTimes()

	s.handler = NewAuthHandler(s.mockAuthService, s.mockJWTMiddleware)

	gin.SetMode(gin.TestMode)
	s.router = gin.New()
	s.handler.SetupRoutes(s.router)
}

func (s *AuthHandlerSuite) TearDownTest() {
	s.ctrl.Finish()
}

func TestAuthHandlerSuite(t *testing.T) {
	suite.Run(t, new(AuthHandlerSuite))
}

func (s *AuthHandlerSuite) TestRegister() {
	user := &entities.User{
		ID:       "1",
		Username: "testuser",
		Email:    "test@example.com",
		Role:     entities.Developer,
		Scopes:   5,
	}

	s.mockAuthService.EXPECT().
		Register("testuser", "password123", "test@example.com", entities.Developer, gomock.Any()).
		Return(user, nil)

	reqBody := dto.RegisterRequest{
		Username: "testuser",
		Password: "password123",
		Email:    "test@example.com",
		Role:     entities.Developer,
	}
	jsonData, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusCreated, w.Code)

	var response dto.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.True(response.Success)
	s.Equal("REGISTER_SUCCESS", response.Code)
}

func (s *AuthHandlerSuite) TestRegisterInvalidRequestBody() {
	req := httptest.NewRequest("POST", "/auth/register", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code)

	var response dto.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.NotEmpty(response.Error)
}

func (s *AuthHandlerSuite) TestRegisterServiceError() {
	s.mockAuthService.EXPECT().
		Register("testuser", "password123", "test@example.com", entities.Developer, gomock.Any()).
		Return((*entities.User)(nil), errors.New("registration failed"))

	reqBody := dto.RegisterRequest{
		Username: "testuser",
		Password: "password123",
		Email:    "test@example.com",
		Role:     entities.Developer,
	}
	jsonData, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusInternalServerError, w.Code)

	var response dto.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal("registration failed", response.Error)
}

func (s *AuthHandlerSuite) TestLogin() {
	accessToken := "test_access_token"

	s.mockAuthService.EXPECT().
		Login(gomock.Any(), "testuser", "password123").
		Return(accessToken, nil)

	reqBody := dto.LoginRequest{
		Username: "testuser",
		Password: "password123",
	}
	jsonData, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusOK, w.Code)

	var response dto.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.True(response.Success)
	s.Equal("LOGIN_SUCCESS", response.Code)

	raw, err := json.Marshal(response.Data)
	s.NoError(err)

	var data dto.LoginResponse
	err = json.Unmarshal(raw, &data)
	s.NoError(err)
	s.Equal(accessToken, data.AccessToken)
}

func (s *AuthHandlerSuite) TestLoginInvalidRequestBody() {
	req := httptest.NewRequest("POST", "/auth/login", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusBadRequest, w.Code)

	var response dto.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.NotEmpty(response.Error)
}

func (s *AuthHandlerSuite) TestLoginServiceError() {
	s.mockAuthService.EXPECT().
		Login(gomock.Any(), "testuser", "wrongpassword").
		Return("", errors.New("service error"))

	reqBody := dto.LoginRequest{
		Username: "testuser",
		Password: "wrongpassword",
	}
	jsonData, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusInternalServerError, w.Code)

	var response dto.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal("service error", response.Error)
}

func (s *AuthHandlerSuite) TestUpdatePassword() {
	s.mockAuthService.EXPECT().
		UpdatePassword(gomock.Any(), "test-user-id", "oldpassword", "newpassword").
		Return(nil)

	reqBody := dto.UpdatePasswordRequest{
		CurrentPassword: "oldpassword",
		NewPassword:     "newpassword",
	}
	jsonData, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("PUT", "/auth/update/password/test-user-id", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusOK, w.Code)

	var response dto.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.True(response.Success)
	s.Equal("PASSWORD_UPDATE_SUCCESS", response.Code)
}

func (s *AuthHandlerSuite) TestUpdatePasswordInvalidRequestBody() {
	req := httptest.NewRequest("PUT", "/auth/update/password/test-user-id", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusBadRequest, w.Code)

	var response dto.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.NotEmpty(response.Error)
}

func (s *AuthHandlerSuite) TestUpdatePasswordServiceError() {
	s.mockAuthService.EXPECT().
		UpdatePassword(gomock.Any(), "test-user-id", "oldpassword", "newpassword").
		Return(errors.New("service error"))

	reqBody := dto.UpdatePasswordRequest{
		CurrentPassword: "oldpassword",
		NewPassword:     "newpassword",
	}
	jsonData, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("PUT", "/auth/update/password/test-user-id", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusInternalServerError, w.Code)

	var response dto.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal("service error", response.Error)
}

func (s *AuthHandlerSuite) TestRefreshAccessToken() {
	accessToken := "test-access-token"
	s.mockAuthService.EXPECT().RefreshAccessToken(gomock.Any(), "test-user-id").Return(accessToken, nil)

	req := httptest.NewRequest("POST", "/auth/refresh/test-user-id", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusOK, w.Code)

	var response dto.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.True(response.Success)
	s.Equal("REFRESH_SUCCESS", response.Code)

	raw, err := json.Marshal(response.Data)
	s.NoError(err)

	var data dto.LoginResponse
	err = json.Unmarshal(raw, &data)
	s.NoError(err)
	s.Equal(accessToken, data.AccessToken)
}

func (s *AuthHandlerSuite) TestRefreshAccessTokenServiceError() {
	s.mockAuthService.EXPECT().RefreshAccessToken(gomock.Any(), "test-user-id").Return("", errors.New("service error"))

	req := httptest.NewRequest("POST", "/auth/refresh/test-user-id", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusInternalServerError, w.Code)

	var response dto.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal("service error", response.Error)
}
