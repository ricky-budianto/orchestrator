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

func CreateWorkflowAudit(metadata model.WorkflowAuditMetadata) cesresponse.Response {
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
	err := dal.CreateWorkflowAudit(context.Background(), &requestData)
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

func UpdateWorkflowAudit(metadata model.WorkflowAuditMetadata) cesresponse.Response {
	logging := ceslogger.Logger{}

	response := cesresponse.Response{
		StatusCode: utils.HTTPBadRequest,
	}

	// PREPARE REQUEST DATA
	requestData := metadata.Body
	requestData.ID = metadata.WorkflowAuditQueryParams.ID
	requestData.UpdatedAt, _ = utils.TimestampNow(false, config.AppConfig.FormatTimestamp)

	// UPDATE DATABASE
	err := dal.UpdateWorkflowAudit(context.Background(), &requestData)
	if err != nil {
		logging.LogError("update data to database using dal", err)
		errorMessage := helper.GenerateError(err.Error(), metadata.Header["locale"], nil)
		response.ResponseData = utils.GenerateJsonResponse(false, errorMessage, nil, err.Error(), nil)
		return response
	}

	response.StatusCode = utils.HTTPOk
	response.ResponseData = utils.GenerateJsonResponse(true, "", requestData, "", nil)
	return response
}

func ListWorkflowAudit(metadata model.WorkflowAuditMetadata) cesresponse.Response {
	logging := ceslogger.Logger{}

	response := cesresponse.Response{
		StatusCode: utils.HTTPBadRequest,
	}

	WorkflowAudits := &[]model.WorkflowAudit{}
	var totalData int64

	queryResult := dal.ListWorkflowAudit(context.Background(), metadata.WorkflowAuditQueryParams, WorkflowAudits, &totalData)
	if queryResult != nil {
		logging.LogError("list customer type data from dal", queryResult.Error)
		errorMessage := helper.GenerateError(cesresponse.RCBadRequest, metadata.Header["locale"], nil)
		response.ResponseData = utils.GenerateJsonResponse(false, errorMessage, nil, cesresponse.RCBadRequest, nil)
		return response
	}

	response.StatusCode = utils.HTTPOk
	response.ResponseData = utils.GenerateJsonResponse(true, "", WorkflowAudits, nil, &totalData)
	return response
}

func GetWorkflowAudit(metadata model.WorkflowAuditMetadata) cesresponse.Response {
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

	var WorkflowAudit model.WorkflowAudit
	WorkflowAudit.ID = metaID

	queryResult := dal.GetWorkflowAudit(context.Background(), &WorkflowAudit)
	if queryResult != nil {
		logging.LogError("get customer type data from dal", queryResult.Error)
		errorMessage := helper.GenerateError(cesresponse.RCDataNotFound, metadata.Header["locale"], nil)
		response.ResponseData = utils.GenerateJsonResponse(false, errorMessage, nil, cesresponse.RCDataNotFound, nil)
		return response
	}

	response.StatusCode = utils.HTTPOk
	response.ResponseData = utils.GenerateJsonResponse(true, "", WorkflowAudit, nil, nil)
	return response
}
