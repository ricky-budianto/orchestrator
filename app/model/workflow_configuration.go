package model

import "gorm.io/gorm"

type WorkflowConfiguration struct {
	ID            string `json:"id" gorm:"type:text"`
	Name          string `json:"name"`
	Path          string `json:"path"`
	Method        string `json:"method"`
	Descriptions  string `json:"descriptions"`
	Configuration string `json:"configuration"  gorm:"type:text"`
	Audit
}

type WorkflowConfigurationHTTPQueryParameter struct {
	HTTPQueryParameter
	ID string `json:"id" query:"id"`
}

type WorkflowConfigurationMetadata struct {
	GofiberMetadata
	Body                             WorkflowConfiguration
	WorkflowConfigurationQueryParams WorkflowConfigurationHTTPQueryParameter
}

func (r *WorkflowConfiguration) BeforeUpdate(tx *gorm.DB) (err error) {
	oldData := &WorkflowConfiguration{}
	oldData.ID = r.ID
	if err = tx.First(oldData).Error; err != nil {
		return err
	}

	tx.Statement.SetColumn("revision_number", oldData.RevisionNumber+1)

	return
}
