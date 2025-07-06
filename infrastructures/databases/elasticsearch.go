package databases

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
)

type ServerStatus struct {
	ServerID    string    `json:"server_id"`
	Status      string    `json:"status"`
	Uptime      int64     `json:"uptime"`
	LastUpdated time.Time `json:"last_updated"`
}

func ConnectESDb() *elasticsearch.Client {
	cfg := elasticsearch.Config{
		Addresses: []string{"http://localhost:9200"},
	}
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatalf("Error creating Elasticsearch client: %s", err)
	}
	return es
}

func SmartBulkUpdateServerStatuses(esClient *elasticsearch.Client, statusList []ServerStatus) error {
	var buf bytes.Buffer
	indexName := "server_status"

	var ids []string
	for _, status := range statusList {
		ids = append(ids, status.ServerID)
	}

	existingDocs, err := GetMultipleServerDocs(esClient, ids)
	if err != nil {
		return fmt.Errorf("mget failed: %w", err)
	}

	for _, status := range statusList {
		old, exists := existingDocs[status.ServerID]
		if !exists {
			meta := map[string]map[string]string{
				"index": {
					"_index": indexName,
					"_id":    status.ServerID,
				},
			}
			metaLine, _ := json.Marshal(meta)
			buf.Write(metaLine)
			buf.WriteByte('\n')

			docLine, _ := json.Marshal(status)
			buf.Write(docLine)
			buf.WriteByte('\n')
			continue
		}

		if old.Status == status.Status {
			newUptime := old.Uptime + status.Uptime
			update := map[string]interface{}{
				"doc": map[string]interface{}{
					"uptime":       newUptime,
					"last_updated": time.Now(),
				},
			}
			meta := map[string]map[string]string{
				"update": {
					"_index": indexName,
					"_id":    status.ServerID,
				},
			}
			metaLine, _ := json.Marshal(meta)
			docLine, _ := json.Marshal(update)

			buf.Write(metaLine)
			buf.WriteByte('\n')
			buf.Write(docLine)
			buf.WriteByte('\n')
		} else {
			newDoc := ServerStatus{
				ServerID:    status.ServerID,
				Status:      status.Status,
				Uptime:      status.Uptime,
				LastUpdated: time.Now(),
			}
			meta := map[string]map[string]string{
				"index": {
					"_index": indexName,
					"_id":    status.ServerID,
				},
			}
			metaLine, _ := json.Marshal(meta)
			docLine, _ := json.Marshal(newDoc)

			buf.Write(metaLine)
			buf.WriteByte('\n')
			buf.Write(docLine)
			buf.WriteByte('\n')
		}
	}

	res, err := esClient.Bulk(bytes.NewReader(buf.Bytes()))
	if err != nil {
		return fmt.Errorf("bulk error: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("bulk response error: %s", res.String())
	}
	return nil
}

func GetMultipleServerDocs(esClient *elasticsearch.Client, ids []string) (map[string]ServerStatus, error) {
	body := map[string]interface{}{
		"ids": ids,
	}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return nil, err
	}

	res, err := esClient.Mget(bytes.NewReader(buf.Bytes()), esClient.Mget.WithIndex("server_status"))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("mget error: %s", res.String())
	}

	var parsed struct {
		Docs []struct {
			ID     string       `json:"_id"`
			Found  bool         `json:"found"`
			Source ServerStatus `json:"_source"`
		} `json:"docs"`
	}
	if err := json.NewDecoder(res.Body).Decode(&parsed); err != nil {
		return nil, err
	}

	results := make(map[string]ServerStatus)
	for _, doc := range parsed.Docs {
		if doc.Found {
			results[doc.ID] = doc.Source
		}
	}
	return results, nil
}
