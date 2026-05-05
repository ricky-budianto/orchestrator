package servicev1
import "context"

import (
	"github.com/soluixdeveloper/ces-orchestratorService/app/function/dal"
	"github.com/soluixdeveloper/ces-orchestratorService/app/helper"
	"github.com/soluixdeveloper/ces-orchestratorService/app/model"
	"github.com/soluixdeveloper/ces-orchestratorService/config"
	"github.com/soluixdeveloper/ces-orchestratorService/config/utils"
	ceslogger "github.com/soluixdeveloper/ces-utilities/v2/ceslogger"
	cesresponse "github.com/soluixdeveloper/ces-utilities/v2/cesresponse"
)

func CreateWorkflowHistory(metadata model.WorkflowHistoryMetadata) cesresponse.Response {
	logging := ceslogger.Logger{}

	response := cesresponse.Response{
		StatusCode: utils.HTTPBadRequest,
	}

	// PREPARE REQUEST DATA
	requestData := metadata.Body
	requestData.CreatedAt, _ = utils.TimestampNow(false, config.AppConfig.FormatTimestamp)
	requestData.UpdateByID = requestData.CreatedByID
	requestData.UpdateByName = requestData.CreatedByName
	requestData.UpdatedAt, _ = utils.TimestampNow(false, config.AppConfig.FormatTimestamp)
	requestData.RevisionNumber = 1

	// CREATE DATA TO DATABASE
	err := dal.CreateWorkflowHistory(context.Background(), &requestData)
	if err != nil {
		logging.LogError("create data to database using dal", err)
		errorMessage := helper.GenerateError(err.Error(), metadata.Header["locale"], nil)
		response.ResponseData = utils.GenerateJsonResponse(false, errorMessage, nil, err.Error(), nil)
		return response
	}

	response.StatusCode = utils.HTTPOk
	response.ResponseData = utils.GenerateJsonResponse(true, "", requestData, "", nil)
	return response
}

func UpdateWorkflowHistory(metadata model.WorkflowHistoryMetadata) cesresponse.Response {
	logging := ceslogger.Logger{}

	response := cesresponse.Response{
		StatusCode: utils.HTTPBadRequest,
	}

	// VALIDATE ID FROM QUERY PARAMS
	workflowHistoryId := metadata.WorkflowHistoryQueryParams.Id
	if workflowHistoryId == "" {
		errorMessage := helper.GenerateError(cesresponse.RCMissingParameter, metadata.Header["locale"], nil)
		response.ResponseData = utils.GenerateJsonResponse(false, errorMessage, nil, cesresponse.RCMissingParameter, nil)
		return response
	}

	// GET EXISTING RECORD TO VERIFY IT EXISTS AND PRESERVE AUDIT TRAIL
	var existingData model.WorkflowHistory
	existingData.Id = workflowHistoryId
	err := dal.GetWorkflowHistory(context.Background(), &existingData)
	if err != nil {
		logging.LogError("get existing workflow history", err)
		errorMessage := helper.GenerateError(cesresponse.RCDataNotFound, metadata.Header["locale"], nil)
		response.ResponseData = utils.GenerateJsonResponse(false, errorMessage, nil, cesresponse.RCDataNotFound, nil)
		return response
	}

	// PREPARE UPDATE DATA - ONLY ALLOW SPECIFIC FIELDS TO BE UPDATED
	requestData := metadata.Body
	updateData := model.WorkflowHistory{
		// Only update business fields, protect audit fields
		Status:         requestData.Status,
		Request:        requestData.Request,
		Response:       requestData.Response,
		AdditionalInfo: requestData.AdditionalInfo,
	}

	// Preserve and increment audit fields
	updateData.UpdatedAt, _ = utils.TimestampNow(false, config.AppConfig.FormatTimestamp)
	updateData.UpdateByID = requestData.UpdateByID
	updateData.UpdateByName = requestData.UpdateByName
	updateData.RevisionNumber = existingData.RevisionNumber + 1

	// VALIDATE BUSINESS RULES
	if updateData.Status != "" {
		validStatuses := []string{"pending", "in_progress", "completed", "failed", "cancelled"}
		isValid := false
		for _, validStatus := range validStatuses {
			if updateData.Status == validStatus {
				isValid = true
				break
			}
		}
		if !isValid {
			errorMessage := helper.GenerateError("invalid status value", metadata.Header["locale"], nil)
			response.ResponseData = utils.GenerateJsonResponse(false, errorMessage, nil, "INVALID_STATUS", nil)
			return response
		}
	}

	// UPDATE DATABASE WITH EXPLICIT ID
	err = dal.UpdateWorkflowHistory(context.Background(), workflowHistoryId, &updateData)
	if err != nil {
		logging.LogError("update data to database using dal", err)
		errorMessage := helper.GenerateError(err.Error(), metadata.Header["locale"], nil)
		response.ResponseData = utils.GenerateJsonResponse(false, errorMessage, nil, err.Error(), nil)
		return response
	}

	response.StatusCode = utils.HTTPOk
	response.ResponseData = utils.GenerateJsonResponse(true, "", updateData, "", nil)
	return response
}

func ListWorkflowHistory(metadata model.WorkflowHistoryMetadata) cesresponse.Response {
	logging := ceslogger.Logger{}

	response := cesresponse.Response{
		StatusCode: utils.HTTPBadRequest,
	}

	WorkflowHistorys := &[]model.WorkflowHistory{}
	var totalData int64

	queryResult := dal.ListWorkflowHistory(context.Background(), metadata.WorkflowHistoryQueryParams, WorkflowHistorys, &totalData)
	if queryResult != nil {
		logging.LogError("list customer type data from dal", queryResult.Error)
		errorMessage := helper.GenerateError(cesresponse.RCBadRequest, metadata.Header["locale"], nil)
		response.ResponseData = utils.GenerateJsonResponse(false, errorMessage, nil, cesresponse.RCBadRequest, nil)
		return response
	}

	response.StatusCode = utils.HTTPOk
	response.ResponseData = utils.GenerateJsonResponse(true, "", WorkflowHistorys, nil, &totalData)
	return response
}

func GetWorkflowHistory(metadata model.WorkflowHistoryMetadata) cesresponse.Response {
	logging := ceslogger.Logger{}

	response := cesresponse.Response{
		StatusCode: utils.HTTPBadRequest,
	}

	metaID := metadata.Params["id"]
	if metaID == "" {
		errorMessage := helper.GenerateError(cesresponse.RCMissingParameter, metadata.Header["locale"], nil)
		response.ResponseData = utils.GenerateJsonResponse(false, errorMessage, nil, cesresponse.RCMissingParameter, nil)
		return response
	}

	var WorkflowHistory model.WorkflowHistory
	WorkflowHistory.Id = metaID

	queryResult := dal.GetWorkflowHistory(context.Background(), &WorkflowHistory)
	if queryResult != nil {
		logging.LogError("get customer type data from dal", queryResult.Error)
		errorMessage := helper.GenerateError(cesresponse.RCDataNotFound, metadata.Header["locale"], nil)
		response.ResponseData = utils.GenerateJsonResponse(false, errorMessage, nil, cesresponse.RCDataNotFound, nil)
		return response
	}

	response.StatusCode = utils.HTTPOk
	response.ResponseData = utils.GenerateJsonResponse(true, "", WorkflowHistory, nil, nil)
	return response
}
