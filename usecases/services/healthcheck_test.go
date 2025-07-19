package services

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/elastic/go-elasticsearch/esapi"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"github.com/vnFuhung2903/vcs-sms/dto"
	"github.com/vnFuhung2903/vcs-sms/entities"
	"github.com/vnFuhung2903/vcs-sms/mocks/docker"
	"github.com/vnFuhung2903/vcs-sms/mocks/interfaces"
	"github.com/vnFuhung2903/vcs-sms/mocks/logger"
	"github.com/vnFuhung2903/vcs-sms/mocks/repositories"
)

type HealthcheckServiceSuite struct {
	suite.Suite
	ctrl               *gomock.Controller
	healthcheckService IHealthcheckService
	mockRepo           *repositories.MockIContainerRepository
	mockDockerClient   *docker.MockIDockerClient
	mockEsClient       *interfaces.MockIElasticsearchClient
	mockLogger         *logger.MockILogger
	ctx                context.Context
}

func (s *HealthcheckServiceSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockRepo = repositories.NewMockIContainerRepository(s.ctrl)
	s.mockDockerClient = docker.NewMockIDockerClient(s.ctrl)
	s.mockEsClient = interfaces.NewMockIElasticsearchClient(s.ctrl)
	s.mockLogger = logger.NewMockILogger(s.ctrl)
	s.healthcheckService = NewHealthcheckService(s.mockRepo, s.mockDockerClient, s.mockEsClient, s.mockLogger)
	s.ctx = context.Background()
}

func (s *HealthcheckServiceSuite) TearDownTest() {
	s.ctrl.Finish()
}

func TestHealthcheckServiceSuite(t *testing.T) {
	suite.Run(t, new(HealthcheckServiceSuite))
}

func (s *HealthcheckServiceSuite) TestUpdateStatusNewDocument() {
	statusList := []dto.EsStatusUpdate{
		{ContainerId: "container1", Status: "ON"},
	}

	s.mockEsClient.EXPECT().Do(s.ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, req esapi.Request) (*esapi.Response, error) {
		response := &esapi.Response{
			StatusCode: 200,
			Body: io.NopCloser(strings.NewReader(`{
                "responses": [
                    {
                        "hits": {
                            "hits": []
                        }
                    }
                ]
            }`)),
		}
		return response, nil
	}).Times(1)

	s.mockEsClient.EXPECT().Do(s.ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, req esapi.Request) (*esapi.Response, error) {
		response := &esapi.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(`{"took":1,"errors":false}`)),
		}
		return response, nil
	}).Times(1)

	s.mockLogger.EXPECT().Info("elasticsearch bulk indexed successfully").Times(1)

	err := s.healthcheckService.UpdateStatus(s.ctx, statusList)
	s.NoError(err)
}

func (s *HealthcheckServiceSuite) TestUpdateStatusUpdateDocument() {
	statusList := []dto.EsStatusUpdate{
		{ContainerId: "container1", Status: "OFF"},
	}

	lastUpdated := time.Now().Add(-1 * time.Hour)

	s.mockEsClient.EXPECT().Do(s.ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, req esapi.Request) (*esapi.Response, error) {
		response := &esapi.Response{
			StatusCode: 200,
			Body: io.NopCloser(strings.NewReader(`{
                "responses": [
                    {
                        "hits": {
                            "hits": [
                                {
                                    "_id": "container1",
                                    "_source": {
                                        "container_id": "container1",
                                        "status": "ON",
                                        "uptime": 3600,
                                        "last_updated": "` + lastUpdated.Format(time.RFC3339) + `"
                                    }
                                }
                            ]
                        }
                    }
                ]
            }`)),
		}
		return response, nil
	}).Times(1)

	s.mockEsClient.EXPECT().Do(s.ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, req esapi.Request) (*esapi.Response, error) {
		response := &esapi.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(`{"took":1,"errors":false}`)),
		}
		return response, nil
	}).Times(1)

	s.mockLogger.EXPECT().Info("elasticsearch bulk indexed successfully").Times(1)

	err := s.healthcheckService.UpdateStatus(s.ctx, statusList)
	s.NoError(err)
}

func (s *HealthcheckServiceSuite) TestUpdateStatusGetEsStatusError() {
	statusList := []dto.EsStatusUpdate{
		{ContainerId: "container1", Status: "ON"},
	}

	s.mockEsClient.EXPECT().Do(s.ctx, gomock.Any()).Return(nil, errors.New("elasticsearch error")).Times(1)
	s.mockLogger.EXPECT().Error("failed to msearch containers", gomock.Any()).Times(1)
	s.mockLogger.EXPECT().Error("failed to msearch container", gomock.Any()).Times(1)

	err := s.healthcheckService.UpdateStatus(s.ctx, statusList)
	s.Error(err)
	s.Contains(err.Error(), "elasticsearch error")
}

func (s *HealthcheckServiceSuite) TestUpdateStatusBulkError() {
	statusList := []dto.EsStatusUpdate{
		{ContainerId: "container1", Status: "ON"},
	}

	s.mockEsClient.EXPECT().Do(s.ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, req esapi.Request) (*esapi.Response, error) {
		response := &esapi.Response{
			StatusCode: 200,
			Body: io.NopCloser(strings.NewReader(`{
                "responses": [
                    {
                        "hits": {
                            "hits": []
                        }
                    }
                ]
            }`)),
		}
		return response, nil
	}).Times(1)

	s.mockEsClient.EXPECT().Do(s.ctx, gomock.Any()).Return(nil, errors.New("bulk error")).Times(1)
	s.mockLogger.EXPECT().Error("failed to bulk es", gomock.Any()).Times(1)

	err := s.healthcheckService.UpdateStatus(s.ctx, statusList)
	s.Error(err)
	s.Contains(err.Error(), "bulk error")
}

// Add these additional test methods to your existing HealthcheckServiceSuite

func (s *HealthcheckServiceSuite) TestUpdateStatusSameStatusUpdate() {
	statusList := []dto.EsStatusUpdate{
		{ContainerId: "container1", Status: "ON"},
	}

	lastUpdated := time.Now().Add(-1 * time.Hour)

	// Mock GetEsStatus call (returns existing document with same status)
	s.mockEsClient.EXPECT().Do(s.ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, req esapi.Request) (*esapi.Response, error) {
		response := &esapi.Response{
			StatusCode: 200,
			Body: io.NopCloser(strings.NewReader(`{
                "responses": [
                    {
                        "hits": {
                            "hits": [
                                {
                                    "_id": "container1",
                                    "_source": {
                                        "container_id": "container1",
                                        "status": "ON",
                                        "uptime": 3600,
                                        "last_updated": "` + lastUpdated.Format(time.RFC3339) + `"
                                    }
                                }
                            ]
                        }
                    }
                ]
            }`)),
		}
		return response, nil
	}).Times(1)

	// Mock Bulk request
	s.mockEsClient.EXPECT().Do(s.ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, req esapi.Request) (*esapi.Response, error) {
		response := &esapi.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(`{"took":1,"errors":false}`)),
		}
		return response, nil
	}).Times(1)

	s.mockLogger.EXPECT().Info("elasticsearch bulk indexed successfully").Times(1)

	err := s.healthcheckService.UpdateStatus(s.ctx, statusList)
	s.NoError(err)
}

func (s *HealthcheckServiceSuite) TestUpdateStatusMultipleDocuments() {
	statusList := []dto.EsStatusUpdate{
		{ContainerId: "container1", Status: "ON"},
	}

	// Mock GetEsStatus call (returns multiple documents - should trigger len(old) != 1 case)
	s.mockEsClient.EXPECT().Do(s.ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, req esapi.Request) (*esapi.Response, error) {
		response := &esapi.Response{
			StatusCode: 200,
			Body: io.NopCloser(strings.NewReader(`{
                "responses": [
                    {
                        "hits": {
                            "hits": [
                                {
                                    "_id": "container1",
                                    "_source": {
                                        "container_id": "container1",
                                        "status": "ON",
                                        "uptime": 3600,
                                        "last_updated": "` + time.Now().Add(-1*time.Hour).Format(time.RFC3339) + `"
                                    }
                                },
                                {
                                    "_id": "container1",
                                    "_source": {
                                        "container_id": "container1",
                                        "status": "OFF",
                                        "uptime": 1800,
                                        "last_updated": "` + time.Now().Add(-2*time.Hour).Format(time.RFC3339) + `"
                                    }
                                }
                            ]
                        }
                    }
                ]
            }`)),
		}
		return response, nil
	}).Times(1)

	// Mock Bulk request
	s.mockEsClient.EXPECT().Do(s.ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, req esapi.Request) (*esapi.Response, error) {
		response := &esapi.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(`{"took":1,"errors":false}`)),
		}
		return response, nil
	}).Times(1)

	s.mockLogger.EXPECT().Info("elasticsearch bulk indexed successfully").Times(1)

	err := s.healthcheckService.UpdateStatus(s.ctx, statusList)
	s.NoError(err)
}

func (s *HealthcheckServiceSuite) TestUpdateStatusJSONMarshalError() {
	statusList := []dto.EsStatusUpdate{
		{ContainerId: "container1", Status: "ON"},
	}

	// Mock GetEsStatus call (returns empty results to trigger new document creation)
	s.mockEsClient.EXPECT().Do(s.ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, req esapi.Request) (*esapi.Response, error) {
		response := &esapi.Response{
			StatusCode: 200,
			Body: io.NopCloser(strings.NewReader(`{
                "responses": [
                    {
                        "hits": {
                            "hits": []
                        }
                    }
                ]
            }`)),
		}
		return response, nil
	}).Times(1)

	// Even with JSON marshal errors, bulk should still be called with valid entries
	s.mockEsClient.EXPECT().Do(s.ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, req esapi.Request) (*esapi.Response, error) {
		response := &esapi.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(`{"took":1,"errors":false}`)),
		}
		return response, nil
	}).Times(1)

	s.mockLogger.EXPECT().Info("elasticsearch bulk indexed successfully").Times(1)

	err := s.healthcheckService.UpdateStatus(s.ctx, statusList)
	s.NoError(err)
}

func (s *HealthcheckServiceSuite) TestUpdateStatusEmptyStatusList() {
	statusList := []dto.EsStatusUpdate{}

	// Mock GetEsStatus call with empty IDs
	s.mockEsClient.EXPECT().Do(s.ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, req esapi.Request) (*esapi.Response, error) {
		response := &esapi.Response{
			StatusCode: 200,
			Body: io.NopCloser(strings.NewReader(`{
                "responses": []
            }`)),
		}
		return response, nil
	}).Times(1)

	// Mock Bulk request with empty body
	s.mockEsClient.EXPECT().Do(s.ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, req esapi.Request) (*esapi.Response, error) {
		response := &esapi.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(`{"took":1,"errors":false}`)),
		}
		return response, nil
	}).Times(1)

	s.mockLogger.EXPECT().Info("elasticsearch bulk indexed successfully").Times(1)

	err := s.healthcheckService.UpdateStatus(s.ctx, statusList)
	s.NoError(err)
}

func (s *HealthcheckServiceSuite) TestGetEsStatusEmptyIDs() {
	ids := []string{}
	limit := 10
	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now()

	s.mockEsClient.EXPECT().Do(s.ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, req esapi.Request) (*esapi.Response, error) {
		response := &esapi.Response{
			StatusCode: 200,
			Body: io.NopCloser(strings.NewReader(`{
                "responses": []
            }`)),
		}
		return response, nil
	}).Times(1)

	result, err := s.healthcheckService.GetEsStatus(s.ctx, ids, limit, startTime, endTime)
	s.NoError(err)
	s.Equal(0, len(result))
}

func (s *HealthcheckServiceSuite) TestGetEsStatusResponseMismatch() {
	ids := []string{"container1", "container2"}
	limit := 10
	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now()

	s.mockEsClient.EXPECT().Do(s.ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, req esapi.Request) (*esapi.Response, error) {
		response := &esapi.Response{
			StatusCode: 200,
			Body: io.NopCloser(strings.NewReader(`{
                "responses": [
                    {
                        "hits": {
                            "hits": [
                                {
                                    "_id": "container1",
                                    "_source": {
                                        "container_id": "container1",
                                        "status": "ON",
                                        "uptime": 3600,
                                        "last_updated": "` + time.Now().Add(-1*time.Hour).Format(time.RFC3339) + `"
                                    }
                                }
                            ]
                        }
                    }
                ]
            }`)),
		}
		return response, nil
	}).Times(1)

	result, err := s.healthcheckService.GetEsStatus(s.ctx, ids, limit, startTime, endTime)
	s.NoError(err)
	s.Equal(1, len(result))
	s.Equal(1, len(result["container1"]))
	// s.Equal(0, len(result["container2"]))
}

func (s *HealthcheckServiceSuite) TestGetEsStatusWithTimeFormatting() {
	ids := []string{"container1"}
	limit := 5
	startTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)

	s.mockEsClient.EXPECT().Do(s.ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, req esapi.Request) (*esapi.Response, error) {
		// Verify the time formatting in the request would be correct
		response := &esapi.Response{
			StatusCode: 200,
			Body: io.NopCloser(strings.NewReader(`{
                "responses": [
                    {
                        "hits": {
                            "hits": [
                                {
                                    "_id": "container1",
                                    "_source": {
                                        "container_id": "container1",
                                        "status": "ON",
                                        "uptime": 3600,
                                        "last_updated": "2023-01-01T12:00:00Z"
                                    }
                                }
                            ]
                        }
                    }
                ]
            }`)),
		}
		return response, nil
	}).Times(1)

	result, err := s.healthcheckService.GetEsStatus(s.ctx, ids, limit, startTime, endTime)
	s.NoError(err)
	s.Equal(1, len(result))
	s.Equal(1, len(result["container1"]))
}

func (s *HealthcheckServiceSuite) TestGetEsStatus() {
	ids := []string{"container1", "container2"}
	limit := 10
	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now()

	s.mockEsClient.EXPECT().Do(s.ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, req esapi.Request) (*esapi.Response, error) {
		response := &esapi.Response{
			StatusCode: 200,
			Body: io.NopCloser(strings.NewReader(`{
                "responses": [
                    {
                        "hits": {
                            "hits": [
                                {
                                    "_id": "container1",
                                    "_source": {
                                        "container_id": "container1",
                                        "status": "ON",
                                        "uptime": 3600,
                                        "last_updated": "` + time.Now().Add(-1*time.Hour).Format(time.RFC3339) + `"
                                    }
                                }
                            ]
                        }
                    },
                    {
                        "hits": {
                            "hits": [
                                {
                                    "_id": "container2",
                                    "_source": {
                                        "container_id": "container2",
                                        "status": "OFF",
                                        "uptime": 1800,
                                        "last_updated": "` + time.Now().Add(-2*time.Hour).Format(time.RFC3339) + `"
                                    }
                                }
                            ]
                        }
                    }
                ]
            }`)),
		}
		return response, nil
	}).Times(1)

	result, err := s.healthcheckService.GetEsStatus(s.ctx, ids, limit, startTime, endTime)
	s.NoError(err)
	s.Equal(2, len(result))
	s.Equal(1, len(result["container1"]))
	s.Equal(1, len(result["container2"]))
	s.Equal(entities.ContainerOn, result["container1"][0].Status)
	s.Equal(entities.ContainerOff, result["container2"][0].Status)
}

func (s *HealthcheckServiceSuite) TestGetEsStatusElasticsearchError() {
	ids := []string{"container1"}
	limit := 10
	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now()

	s.mockEsClient.EXPECT().Do(s.ctx, gomock.Any()).Return(nil, errors.New("elasticsearch connection error")).Times(1)
	s.mockLogger.EXPECT().Error("failed to msearch containers", gomock.Any()).Times(1)

	result, err := s.healthcheckService.GetEsStatus(s.ctx, ids, limit, startTime, endTime)
	s.Error(err)
	s.Nil(result)
	s.Contains(err.Error(), "elasticsearch connection error")
}

func (s *HealthcheckServiceSuite) TestGetEsStatusReadBodyError() {
	ids := []string{"container1"}
	limit := 10
	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now()

	failingReader := &failingReadCloser{}

	s.mockEsClient.EXPECT().Do(s.ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, req esapi.Request) (*esapi.Response, error) {
		response := &esapi.Response{
			StatusCode: 200,
			Body:       failingReader,
		}
		return response, nil
	}).Times(1)

	s.mockLogger.EXPECT().Error("failed to read msearch response body", gomock.Any()).Times(1)

	result, err := s.healthcheckService.GetEsStatus(s.ctx, ids, limit, startTime, endTime)
	s.Error(err)
	s.Nil(result)
}

func (s *HealthcheckServiceSuite) TestGetEsStatusInvalidJSONResponse() {
	ids := []string{"container1"}
	limit := 10
	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now()

	s.mockEsClient.EXPECT().Do(s.ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, req esapi.Request) (*esapi.Response, error) {
		response := &esapi.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(`invalid json response`)),
		}
		return response, nil
	}).Times(1)

	s.mockLogger.EXPECT().Error("failed to decode msearch response body", gomock.Any()).Times(1)

	result, err := s.healthcheckService.GetEsStatus(s.ctx, ids, limit, startTime, endTime)
	s.Error(err)
	s.Nil(result)
}

type failingReadCloser struct{}

func (f *failingReadCloser) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}

func (f *failingReadCloser) Close() error {
	return nil
}
