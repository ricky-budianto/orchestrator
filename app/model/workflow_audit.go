package model

import (
	"time"

	"gorm.io/gorm"
)

type WorkflowAudit struct {
	ID             string         `json:"id" gorm:"type:varchar(36);default:uuid_generate_v4()"`
	ResourceId     string         `json:"resource_id,omitempty"`
	Data           interface{}    `json:"data,omitempty" gorm:"serializer:json"`
	RevisionNumber int            `json:"revision_number,omitempty" gorm:"type:int8;default:1"`
	CreatedAt      time.Time      `json:"created_at"`
	CreatedByID    string         `json:"created_by_id"`
	CreatedByName  string         `json:"created_by_name"`
	UpdatedAt      time.Time      `json:"updated_at"`
	UpdateByID     string         `json:"update_by_id"`
	UpdateByName   string         `json:"update_by_name"`
	DeletedAt      gorm.DeletedAt `json:"deleted_at"`
}

type WorkflowAuditHTTPQueryParameter struct {
	HTTPQueryParameter
	ID             string `json:"id,omitempty" query:"id"`
	RevisionNumber int    `json:"revision_number,omitempty" query:"revision_number"`
	ResourceId     string `json:"resource_id,omitempty" query:"resource_id"`
}

type WorkflowAuditMetadata struct {
	GofiberMetadata
	Body                     WorkflowAudit
	WorkflowAuditQueryParams WorkflowAuditHTTPQueryParameter
}
