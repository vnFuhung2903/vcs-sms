package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/esapi"
	"github.com/vnFuhung2903/vcs-sms/dto"
	"github.com/vnFuhung2903/vcs-sms/interfaces"
	"github.com/vnFuhung2903/vcs-sms/pkg/logger"
	"go.uber.org/zap"
)

type IHealthcheckService interface {
	UpdateStatus(ctx context.Context, statusList []dto.EsStatusUpdate, interval time.Duration) error
	GetEsStatus(ctx context.Context, ids []string, limit int, startTime time.Time, endTime time.Time, order dto.SortOrder) (map[string][]dto.EsStatus, error)
}

type HealthcheckService struct {
	esClient interfaces.IElasticsearchClient
	logger   logger.ILogger
}

func NewHealthcheckService(esClient interfaces.IElasticsearchClient, logger logger.ILogger) IHealthcheckService {
	return &HealthcheckService{
		esClient: esClient,
		logger:   logger,
	}
}

func (s *HealthcheckService) UpdateStatus(ctx context.Context, statusList []dto.EsStatusUpdate, interval time.Duration) error {
	var buf bytes.Buffer
	indexName := "sms_container"

	var ids []string
	for _, status := range statusList {
		ids = append(ids, status.ContainerId)
	}

	endTime := time.Now()
	startTime := endTime.Add(-interval)
	var zeroTime time.Time

	existingDocs, err := s.GetEsStatus(ctx, ids, 1, startTime, endTime, dto.Dsc)
	if err != nil {
		return err
	}
	previousDocs, err := s.GetEsStatus(ctx, ids, 1, zeroTime, startTime, dto.Dsc)
	if err != nil {
		return err
	}

	for _, status := range statusList {
		var (
			meta interface{}
			doc  interface{}
		)

		old := existingDocs[status.ContainerId]
		previous := previousDocs[status.ContainerId]

		switch {
		case len(old) == 0:
			nextCounter := int64(0)
			if len(previous) > 0 {
				nextCounter = previous[0].Counter + 1
			}
			meta = map[string]map[string]string{
				"index": {
					"_index": indexName,
					"_id":    fmt.Sprintf("%s_%d", status.ContainerId, nextCounter),
				},
			}
			doc = dto.EsStatus{
				ContainerId: status.ContainerId,
				Status:      status.Status,
				LastUpdated: endTime,
				Uptime:      int64(interval.Seconds()),
				Counter:     nextCounter,
			}

		case old[0].Status == status.Status:
			meta = map[string]map[string]string{
				"update": {
					"_index": indexName,
					"_id":    fmt.Sprintf("%s_%d", status.ContainerId, old[0].Counter),
				},
			}
			doc = map[string]interface{}{
				"doc": map[string]interface{}{
					"uptime":       old[0].Uptime + int64(endTime.Sub(old[0].LastUpdated).Seconds()),
					"last_updated": endTime,
				},
			}

		default:
			nextCounter := old[0].Counter + 1
			meta = map[string]map[string]string{
				"index": {
					"_index": indexName,
					"_id":    fmt.Sprintf("%s_%d", status.ContainerId, nextCounter),
				},
			}
			doc = dto.EsStatus{
				ContainerId: status.ContainerId,
				Status:      status.Status,
				Uptime:      int64(endTime.Sub(old[0].LastUpdated).Seconds()),
				LastUpdated: endTime,
				Counter:     nextCounter,
			}
		}

		metaLine, err := json.Marshal(meta)
		if err != nil {
			s.logger.Error("failed to marshal meta", zap.Error(err))
			continue
		}

		docLine, err := json.Marshal(doc)
		if err != nil {
			s.logger.Error("failed to marshal doc", zap.Error(err))
			continue
		}

		buf.Write(metaLine)
		buf.WriteByte('\n')
		buf.Write(docLine)
		buf.WriteByte('\n')
	}

	req := esapi.BulkRequest{
		Body: bytes.NewReader(buf.Bytes()),
	}
	res, err := s.esClient.Do(ctx, req)
	if err != nil {
		s.logger.Error("failed to bulk elasticsearch status", zap.Error(err))
		return err
	}
	defer res.Body.Close()
	s.logger.Info("elasticsearch status indexed successfully")
	return nil
}

func (s *HealthcheckService) GetEsStatus(ctx context.Context, ids []string, limit int, startTime time.Time, endTime time.Time, order dto.SortOrder) (map[string][]dto.EsStatus, error) {
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
				map[string]interface{}{"counter": map[string]string{"order": string(order)}},
			},
		}
		queryLine, _ := json.Marshal(query)
		body.Write(queryLine)
		body.WriteByte('\n')
	}

	req := esapi.MsearchRequest{
		Body: strings.NewReader(body.String()),
	}
	res, err := s.esClient.Do(ctx, req)
	if err != nil {
		s.logger.Error("failed to msearch elasticsearch status", zap.Error(err))
		return nil, err
	}
	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		s.logger.Error("failed to read response body", zap.Error(err))
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
		s.logger.Error("failed to decode response body", zap.Error(err))
		return nil, err
	}

	results := make(map[string][]dto.EsStatus)
	for i, response := range parsed.Responses {
		containerId := ids[i]
		for _, hit := range response.Hits.Hits {
			results[containerId] = append(results[containerId], hit.Source)
		}
	}
	s.logger.Info("elasticsearch status retrieved successfully", zap.Int("containers_count", len(results)))
	return results, nil
}
