package model

type WorkflowHistory struct {
	Id                      string      `json:"id" gorm:"type:text"`
	WorkflowConfigurationId string      `json:"workflow_configuration_id"`
	Status                  string      `json:"status"`
	Request                 interface{} `json:"request"  gorm:"serializer:json"`
	Response                interface{} `json:"response"  gorm:"serializer:json"`
	AdditionalInfo          interface{} `json:"additional_info"  gorm:"serializer:json"`
	Audit
}

type WorkflowHistoryHTTPQueryParameter struct {
	HTTPQueryParameter
	Id                      string `json:"id" query:"id"`
	WorkflowConfigurationId string `json:"workflow_configuration_id" query:"workflow_configuration_id"`
	Status                  string `json:"status" query:"status"`
}

type WorkflowHistoryMetadata struct {
	GofiberMetadata
	Body                       WorkflowHistory
	WorkflowHistoryQueryParams WorkflowHistoryHTTPQueryParameter
}

// WorkflowHistoryUpdateRequest is a DTO for update operations to prevent mass assignment
type WorkflowHistoryUpdateRequest struct {
	WorkflowConfigurationId string      `json:"workflow_configuration_id" validate:"omitempty,uuid"`
	Status                  string      `json:"status" validate:"required,oneof=pending in_progress completed failed cancelled"`
	Request                 interface{} `json:"request"`
	Response                interface{} `json:"response"`
	AdditionalInfo          interface{} `json:"additional_info"`
	UpdateByID              string      `json:"update_by_id" validate:"required"`
	UpdateByName            string      `json:"update_by_name" validate:"required"`
}
