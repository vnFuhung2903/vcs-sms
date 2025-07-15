package services

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/vnFuhung2903/vcs-sms/dto"
	"github.com/vnFuhung2903/vcs-sms/entities"
	"github.com/vnFuhung2903/vcs-sms/pkg/docker"
	"github.com/vnFuhung2903/vcs-sms/pkg/logger"
	"github.com/vnFuhung2903/vcs-sms/usecases/repositories"
	"go.uber.org/zap"
)

type IHealthcheckService interface {
	UpdateStatus(ctx context.Context, statusList []dto.EsStatusUpdate) error
	CalculateUptimePercentage(statusList []dto.EsStatus, startTime time.Time, endTime time.Time) float64
	GetEsStatus(ctx context.Context, ids []string, limit int, startTime time.Time, endTime time.Time) (map[string][]dto.EsStatus, error)
}

type HealthcheckService struct {
	containerRepo repositories.IContainerRepository
	esClient      *elasticsearch.Client
	logger        logger.ILogger
}

func NewHealthcheckService(repo repositories.IContainerRepository, dockerClient docker.IDockerClient, esClient *elasticsearch.Client, logger logger.ILogger) IHealthcheckService {
	return &HealthcheckService{
		containerRepo: repo,
		esClient:      esClient,
		logger:        logger,
	}
}

func (s *HealthcheckService) UpdateStatus(ctx context.Context, statusList []dto.EsStatusUpdate) error {
	var buf bytes.Buffer
	indexName := "sms_container"

	var ids []string
	for _, status := range statusList {
		ids = append(ids, status.ContainerId)
	}

	endTime := time.Now()
	startTime := endTime.Add(-24 * time.Hour)
	existingDocs, err := s.GetEsStatus(ctx, ids, 1, startTime, endTime)
	if err != nil {
		s.logger.Error("failed to msearch container", zap.Error(err))
		return err
	}

	var metaLine, docLine []byte

	for _, status := range statusList {
		old := existingDocs[status.ContainerId]
		if len(old) != 1 {
			meta := map[string]map[string]string{
				"index": {
					"_index": indexName,
					"_id":    status.ContainerId,
				},
			}

			metaLine, err = json.Marshal(meta)
			if err != nil {
				s.logger.Error("failed to create json", zap.Error(err))
			}
			docLine, err = json.Marshal(dto.EsStatus{
				ContainerId: status.ContainerId,
				Status:      status.Status,
				LastUpdated: time.Now().UTC().Truncate(time.Second),
				Uptime:      0,
			})
			if err != nil {
				s.logger.Error("failed to create json", zap.Error(err))
			}
		} else if old[0].Status == status.Status {
			update := map[string]interface{}{
				"doc": map[string]interface{}{
					"uptime":       old[0].Uptime + int64(time.Since(old[0].LastUpdated)),
					"last_updated": time.Now().UTC().Truncate(time.Second),
				},
			}
			meta := map[string]map[string]string{
				"update": {
					"_index": indexName,
					"_id":    status.ContainerId,
				},
			}

			metaLine, err = json.Marshal(meta)
			if err != nil {
				s.logger.Error("failed to create json", zap.Error(err))
			}
			docLine, err = json.Marshal(update)
			if err != nil {
				s.logger.Error("failed to create json", zap.Error(err))
			}
		} else {
			newDoc := dto.EsStatus{
				ContainerId: status.ContainerId,
				Status:      status.Status,
				Uptime:      int64(time.Since(old[0].LastUpdated)),
				LastUpdated: time.Now(),
			}
			meta := map[string]map[string]string{
				"index": {
					"_index": indexName,
					"_id":    status.ContainerId,
				},
			}

			metaLine, err = json.Marshal(meta)
			if err != nil {
				s.logger.Error("failed to create json", zap.Error(err))
			}
			docLine, err = json.Marshal(newDoc)
			if err != nil {
				s.logger.Error("failed to create json", zap.Error(err))
			}
		}

		if err == nil {
			buf.Write(metaLine)
			buf.WriteByte('\n')
			buf.Write(docLine)
			buf.WriteByte('\n')
		}
	}

	res, err := s.esClient.Bulk(bytes.NewReader(buf.Bytes()))
	if err != nil {
		s.logger.Error("failed to bulk es", zap.Error(err))
		return err
	}
	defer res.Body.Close()
	s.logger.Info("elasticsearch bulk indexed successfully")
	return nil
}

func (s *HealthcheckService) CalculateUptimePercentage(statusList []dto.EsStatus, startTime time.Time, endTime time.Time) float64 {
	result := 0.0
	if endTime.After(time.Now()) {
		endTime = time.Now()
	}
	prevTime := endTime
	for _, status := range statusList {
		if status.Status == entities.ContainerOn {
			result += prevTime.Sub(status.LastUpdated).Hours()
		} else {
			prevTime = status.LastUpdated
		}
	}
	return result
}

func (s *HealthcheckService) GetEsStatus(ctx context.Context, ids []string, limit int, startTime time.Time, endTime time.Time) (map[string][]dto.EsStatus, error) {
	var body strings.Builder

	for _, id := range ids {
		meta := map[string]string{"index": "sms_container"}
		metaLine, _ := json.Marshal(meta)
		body.Write(metaLine)
		body.WriteByte('\n')

		query := map[string]interface{}{
			"query": map[string]interface{}{
				"bool": map[string]interface{}{
					"must": []interface{}{
						map[string]interface{}{"term": map[string]string{"container_id.keyword": id}},
						map[string]interface{}{
							"range": map[string]interface{}{
								"last_updated": map[string]string{
									"gte": startTime.Format(time.RFC3339),
									"lt":  endTime.Format(time.RFC3339),
								},
							},
						},
					},
				},
			},
			"size": limit,
			"sort": []interface{}{
				map[string]interface{}{"last_updated": map[string]string{"order": "desc"}},
			},
		}
		queryLine, _ := json.Marshal(query)
		body.Write(queryLine)
		body.WriteByte('\n')
	}

	res, err := s.esClient.Msearch(
		strings.NewReader(body.String()),
		s.esClient.Msearch.WithContext(ctx),
	)
	if err != nil {
		s.logger.Error("failed to msearch containers", zap.Error(err))
		return nil, err
	}
	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		s.logger.Error("failed to read msearch response body", zap.Error(err))
		return nil, err
	}

	var parsed struct {
		Responses []struct {
			Hits struct {
				Hits []struct {
					ID     string       `json:"_id"`
					Source dto.EsStatus `json:"_source"`
				} `json:"hits"`
			} `json:"hits"`
		} `json:"responses"`
	}
	if err := json.Unmarshal(bodyBytes, &parsed); err != nil {
		s.logger.Error("failed to decode msearch response body", zap.Error(err))
		return nil, err
	}

	results := make(map[string][]dto.EsStatus)
	for i, response := range parsed.Responses {
		containerId := ids[i]
		for _, hit := range response.Hits.Hits {
			results[containerId] = append(results[containerId], hit.Source)
		}
	}
	return results, nil
}
