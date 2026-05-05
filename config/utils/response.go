package utils

// Response parent object for response
// use this as all purpose response object
type Response struct {
	ResponseData interface{}
	StatusCode   int
}

// JsonResponse json object for response
type JsonResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	ErrorCode interface{} `json:"error_code,omitempty"`
	TotalData *int64      `json:"total_data,omitempty"`
}

// GenerateJsonResponse used to generate json response
func GenerateJsonResponse(success bool, message string, data interface{}, err interface{}, totalData *int64) JsonResponse {
	response := JsonResponse{}

	response.Success = success
	response.Message = message
	response.Data = data
	response.ErrorCode = err
	response.TotalData = totalData

	return response
}
