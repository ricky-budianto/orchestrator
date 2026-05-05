package model

import "time"

type State struct {
	RequestSent      time.Time   `json:"request_sent"`
	ResponseReceived time.Time   `json:"response_received"`
	Success          bool        `json:"success"`
	Message          string      `json:"message"`
	AdditionalInfo   interface{} `json:"additional_info"`
}
type WorkflowState struct {
	Id                  string      `json:"id" gorm:"type:varchar(36);default:uuid_generate_v4()"`
	WorkflowId          string      `json:"workflow_id"`
	WorkflowRequestType string      `json:"workflow_request_type"`
	State               interface{} `json:"state,omitempty" gorm:"serializer:json"`
	Audit
}

type WorkflowStateHTTPQueryParameter struct {
	HTTPQueryParameter
	Id                  string `json:"id,omitempty" query:"id"`
	WorkflowId          string `json:"workflow_id,omitempty" query:"workflow_id"`
	WorkflowRequestType string `json:"workflow_request_type,omitempty" query:"workflow_request_type"`
}

type WorkflowStateMetadata struct {
	GofiberMetadata
	Body                     WorkflowState
	WorkflowStateQueryParams WorkflowStateHTTPQueryParameter
}

type OrchestrationLog struct {
	WorkflowId          string      `json:"workflow_id"`
	Task                string      `json:"task"`
	WorkflowRequestType string      `json:"workflow_request_type"`
	Worker              string      `json:"worker,omitempty"`
	RequestType         string      `json:"request_type,omitempty"`
	Request             interface{} `json:"request" gorm:"serializer:json"`
	Response            interface{} `json:"response" gorm:"serializer:json"`
	VariableCheck       interface{} `json:"variable_check"`
	TargetState         string      `json:"target_state"`
	StartTimestamp      time.Time   `json:"start_timestamp"`
	EndTimestamp        time.Time   `json:"end_timestamp"`
}
