package model

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type ResponseData struct {
	Success      bool        `json:"success"`
	Data         interface{} `json:"data"`
	Message      string      `json:"message"`
	ResponseCode string      `json:"response_code"`
	FullCount    int         `json:"full_count,omitempty"`
	Count        int         `json:"count,omitempty"`
	ErrorCode    interface{} `json:"error_code,omitempty"`
	TotalData    *int64      `json:"total_data,omitempty"`
	StatusCode   int         `json:"status_code,omitempty"`
	ResponseData interface{} `json:"ResponseData,omitempty"`
}

type UpdateData struct {
	OldData interface{} `json:"old_data"`
	NewData interface{} `json:"new_data"`
}

type HTTPRequest struct {
	Context       fiber.Ctx                `json:"Context"`
	QueryParams   map[string][]string      `json:"QueryParams"`
	FormValue     map[string][]interface{} `json:"Formvalue"`
	Params        map[string]string        `json:"Params"`
	Headers       map[string]string        `json:"Headers"`
	Formfile      map[string][]interface{} `json:"Formfile"`
	Skip          string                   `json:"Skip"`
	Path          string                   `json:"Path"`
	Method        string                   `json:"Method"`
	Authorization string                   `json:"Authorization"`
	Limit         string                   `json:"Limit"`
	Tabel         string                   `json:"Tabel"`
	OriginalURL   string                   `json:"OriginalURL"`
	RequestType   string                   `json:"RequestType"`
	RequestId     string                   `json:"RequestId"`
	Body          []byte                   `json:"Body"`
	BodyRaw       []byte                   `json:"BodyRaw"`
}
type HTTPHeader struct {
	Method        string
	Authorization string
	Skip          string
	Limit         string
	Tabel         string
	Action        string
}

type BasicAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Metadata defines model meta data for every protocol, e.g: HTTP, RPC
type Metadata struct {
	Path        string
	Header      map[string]string
	BodyRaw     []byte
	Params      map[string]string
	Body        interface{}
	Headers     map[string]string
	QueryParams map[string][]string
	FormFile    map[string]interface{}
	FormValue   map[string]interface{}
}

type MetadataRabbit struct {
	Path        string
	Header      map[string]string
	BodyRaw     []byte
	Params      map[string]string
	Body        interface{}
	Headers     map[string]string
	QueryParams map[string]interface{}
	FormFile    map[string]interface{}
	FormValue   map[string]interface{}
}

type GofiberMetadataRabbit struct {
	MetadataRabbit
	Context *fiber.Ctx
}

type GofiberMetadata struct {
	Metadata
	Context *fiber.Ctx
}

type MappingAccount struct {
	BankCode  string `json:"bank_code,omitempty" query:"bank_code"`
	SwiftCode string `json:"swift_code,omitempty" query:"swift_code"`
}

type HTTPQueryParameter struct {
	Limit   int    `json:"-" query:"limit"`
	Offset  int    `json:"-" query:"offset"`
	OrderBy string `json:"-" query:"order_by"`
	Desc    string `json:"-" query:"desc"`
	MappingAccount
}

type SystemParameter struct {
	ID           string `json:"id"`
	Group        string `json:"group"`
	Value        string `json:"value"`
	Descriptions string `json:"descriptions"`
}

type Error struct {
	ID            string            `json:"id"`
	Descriptions  map[string]string `json:"descriptions" gorm:"serializer:json"`
	ProblemOwner  string            `json:"problem_owner"`
	SeverityLevel byte              `json:"severity_level"`
	WhatToDo      map[string]string `json:"what_to_do" gorm:"serializer:json"`
}

type PayloadToken struct {
	Sub               string `json:"sub,omitempty"`                //ID
	PreferredUsername string `json:"preferred_username,omitempty"` // username
	Name              string `json:"name"`
}

type MappingTable struct {
	ID          string `json:"id" gorm:"type:varchar(36);default:uuid_generate_v4()"`
	KeyName     string `json:"key_name"`
	InputValue  string `json:"input_value"`
	OutputValue string `json:"output_value"`
}

type Audit struct {
	RevisionNumber int            `json:"revision_number,omitempty" gorm:"type:int8;default:1"`
	CreatedAt      time.Time      `json:"created_at"`
	CreatedByID    string         `json:"created_by_id"`
	CreatedByName  string         `json:"created_by_name"`
	UpdatedAt      time.Time      `json:"updated_at"`
	UpdateByID     string         `json:"update_by_id"`
	UpdateByName   string         `json:"update_by_name"`
	DeletedAt      gorm.DeletedAt `json:"deleted_at"`
}

type JsonResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	ErrorCode interface{} `json:"error_code,omitempty"`
	TotalData *int64      `json:"total_data,omitempty"`
}

func GenerateJsonResponse(success bool, message string, data interface{}, err interface{}, totalData *int64) JsonResponse {
	response := JsonResponse{}

	response.Success = success
	response.Message = message
	response.Data = data
	response.ErrorCode = err
	response.TotalData = totalData

	return response
}
