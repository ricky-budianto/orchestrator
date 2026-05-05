package model

import "time"

type ElasticLogHTTPQueryParameter struct {
	ID                      string    `json:"id,omitempty" query:"id"`
	WorkflowId              string    `json:"workflow_id,omitempty" query:"workflow_id"`
	WofkflowConfigurationID string    `json:"workflow_configuration_id,omitempty" query:"workflow_configuration_id"`
	RequestType             string    `json:"request_type,omitempty" query:"request_type"`
	Worker                  string    `json:"worker,omitempty" query:"worker"`
	StartTimestamp          time.Time `json:"start_timestamp,omitempty" query:"start_timestamp"`
	EndTimestamp            time.Time `json:"end_timestamp,omitempty" query:"end_timestamp"`
	Index                   string    `json:"index,omitempty" query:"index"`
	Task                    string    `json:"task" query:"task"`
}

type ElasticLogMetadata struct {
	GofiberMetadata
	Body                         map[string]interface{}
	ElasticLogHTTPQueryParameter ElasticLogHTTPQueryParameter
}
