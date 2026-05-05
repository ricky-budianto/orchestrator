package elasticlog

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/olivere/elastic/v7"
	"github.com/soluixdeveloper/ces-orchestratorService/app/model"
	"github.com/soluixdeveloper/ces-orchestratorService/config"
	ceslogger "github.com/soluixdeveloper/ces-utilities/v2/ceslogger"
)

var ElasticLog *Logger

// Logger represents an Elasticsearch logger
type Logger struct {
	client *elastic.Client
}

// NewLogger creates a new Elasticsearch logger
func NewLogger() *Logger {
	logging := ceslogger.NewLogger("")
	var client *elastic.Client
	var err error
	if config.ElasticSearchConfig.ElasticSearchAPIKey == "" {
		client, err = elastic.NewClient(
			elastic.SetURL(config.ElasticSearchConfig.ElasticSearchURL+
				":"+
				config.ElasticSearchConfig.ElasticSearchPort),
			elastic.SetSniff(false),
			elastic.SetBasicAuth(config.ElasticSearchConfig.ElasticSearchUsername,
				config.ElasticSearchConfig.ElasticSearchPassword),
		)
	} else {
		url:=config.ElasticSearchConfig.ElasticSearchURL+":"+config.ElasticSearchConfig.ElasticSearchPort
		client, err = elastic.NewClient(
			elastic.SetURL(url),
			elastic.SetSniff(false),
			elastic.SetHeaders(http.Header{"Authorization": []string{"ApiKey " + config.ElasticSearchConfig.ElasticSearchAPIKey}}),
		)
	}
	if err != nil {
		logging.LogError("error initiate elastic log", err.Error())
		return nil
	}

	return &Logger{
		client: client,
	}
}

// Log logs an entry to Elasticsearch
func (l *Logger) Log(entry interface{}, index string) error {
	logging := ceslogger.NewLogger("")
	_, err := l.client.Index().
		Index(index).
		BodyJson(entry).
		Do(context.Background())
	if err != nil {
		logging.LogError("error create elastic log", err.Error())
	}
	return err
}

// Get retrieves a log entry by its ID
func (l *Logger) Get(id string, index string) (*interface{}, error) {
	result, err := l.client.Get().
		Index(index).
		Id(id).
		Do(context.Background())
	if err != nil {
		if elastic.IsNotFound(err) {
			return nil, errors.New("log entry not found")
		}
		return nil, err
	}

	var logEntry interface{}
	err = json.Unmarshal(result.Source, &logEntry)
	if err != nil {
		return nil, err
	}

	return &logEntry, nil
}

// Query retrieves log entries based on specific criteria
func (l *Logger) Query(filter model.ElasticLogHTTPQueryParameter) ([]interface{}, error) {
	boolQuery := elastic.NewBoolQuery()
	// for k, v := range filter {
	// 	boolQuery = boolQuery.Must(elastic.NewTermQuery(k+".keyword", v))
	// }
	if filter.ID != "" {
		boolQuery = boolQuery.Must(elastic.NewTermQuery("id.keyword", filter.ID))
	}
	if filter.WorkflowId != "" {
		boolQuery = boolQuery.Must(elastic.NewTermQuery("workflow_id.keyword", filter.WorkflowId))
	}
	if filter.WofkflowConfigurationID != "" {
		boolQuery = boolQuery.Must(elastic.NewTermQuery("workflow_configuration_id.keyword", filter.WofkflowConfigurationID))
	}
	if filter.RequestType != "" {
		boolQuery = boolQuery.Must(elastic.NewTermQuery("request_type.keyword", filter.RequestType))
	}
	if filter.Worker != "" {
		boolQuery = boolQuery.Must(elastic.NewTermQuery("worker.keyword", filter.Worker))
	}
	if filter.Task != "" {
		boolQuery = boolQuery.Must(elastic.NewTermQuery("task.keyword", filter.Task))
	}
	// Add timestamp range query
	if !filter.StartTimestamp.IsZero() || !filter.EndTimestamp.IsZero() {
		rangeQuery := elastic.NewRangeQuery("log_timestamp")
		if !filter.StartTimestamp.IsZero() {
			rangeQuery = rangeQuery.Gte(filter.StartTimestamp)
		}
		if !filter.EndTimestamp.IsZero() {
			rangeQuery = rangeQuery.Lte(filter.EndTimestamp)
		}
		boolQuery = boolQuery.Filter(rangeQuery)
	}

	searchResult, err := l.client.Search().
		Index(filter.Index).
		Query(boolQuery).
		Do(context.Background())
	if err != nil {
		return nil, err
	}

	var logEntries []interface{}
	for _, hit := range searchResult.Hits.Hits {
		var logEntry interface{}
		err := json.Unmarshal(hit.Source, &logEntry)
		if err != nil {
			return nil, err
		}
		logEntries = append(logEntries, logEntry)
	}

	return logEntries, nil
}
