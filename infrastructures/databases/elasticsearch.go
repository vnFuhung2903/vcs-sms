package databases

import (
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/vnFuhung2903/vcs-sms/pkg/env"
)

type IElasticsearchFactory interface {
	ConnectElasticsearch() (*elasticsearch.Client, error)
}

type elasticsearchFactory struct {
	Address string
}

func NewElasticsearchFactory(env env.ElasticsearchEnv) IElasticsearchFactory {
	return &elasticsearchFactory{Address: env.ElasticsearchAddress}
}

func (f *elasticsearchFactory) ConnectElasticsearch() (*elasticsearch.Client, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{f.Address},
	}
	es, err := elasticsearch.NewClient(cfg)
	return es, err
}
