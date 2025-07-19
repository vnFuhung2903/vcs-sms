package interfaces

import (
	"context"
	"testing"

	"github.com/elastic/go-elasticsearch/esapi"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/stretchr/testify/assert"
)

func TestElasticsearchClient(t *testing.T) {
	cfg := elasticsearch.Config{
		Addresses: []string{"http://localhost:9200"},
	}
	es, err := elasticsearch.NewClient(cfg)
	assert.NoError(t, err)

	esClient := NewElasticsearchClient(es)
	req := esapi.InfoRequest{}
	resp, err := esClient.Do(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, resp)
}
