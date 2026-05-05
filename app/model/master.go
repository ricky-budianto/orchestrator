package model

import (
	"time"

	"gorm.io/gorm"
)

type Master struct {
	ID             string         `json:"id"`
	Value          string         `json:"value,omitempty"`
	Group          string         `json:"group,omitempty"`
	Code           string         `json:"code"`
	RevisionNumber int            `json:"revision_number,omitempty" gorm:"type:uint;default:1"`
	Descriptions   string         `json:"descriptions,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	CreatedByID    string         `json:"created_by_id"`
	CreatedByName  string         `json:"created_by_name"`
	UpdatedAt      time.Time      `json:"updated_at"`
	UpdateByID     string         `json:"update_by_id"`
	UpdateByName   string         `json:"update_by_name"`
	DeletedAt      gorm.DeletedAt `json:"deleted_at"`
}

type MasterHTTPQueryParameter struct {
	HTTPQueryParameter
	ID    string `json:"id,omitempty" query:"id"`
	Code  string `json:"code,omitempty" query:"code"`
	Group string `json:"group,omitempty" query:"group"`
}

type MasterMetadata struct {
	GofiberMetadata
	Body              Master
	MasterQueryParams MasterHTTPQueryParameter
}