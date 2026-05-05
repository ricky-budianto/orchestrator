package model

type Workflow struct {
	ID    string `json:"id" gorm:"type:text"`
	Value string `json:"value,omitempty"`
	Group string `json:"group,omitempty"`
	Code  string `json:"code"`
	Audit
}

type WorkflowHTTPQueryParameter struct {
	HTTPQueryParameter
	ID    string `json:"id,omitempty" query:"id"`
	Code  string `json:"code,omitempty" query:"code"`
	Group string `json:"group,omitempty" query:"group"`
}

type WorkflowMetadata struct {
	GofiberMetadata
	Body                Workflow
	WorkflowQueryParams WorkflowHTTPQueryParameter
}
