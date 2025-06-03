package databases

import (
	"log"

	"github.com/elastic/go-elasticsearch/v8"
)

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
