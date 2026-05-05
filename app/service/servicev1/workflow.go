package servicev1

import (
	"github.com/soluixdeveloper/ces-orchestratorService/app/function/dal"
	"github.com/soluixdeveloper/ces-orchestratorService/app/helper"
	"github.com/soluixdeveloper/ces-orchestratorService/app/model"
	"github.com/soluixdeveloper/ces-orchestratorService/config"
	"github.com/soluixdeveloper/ces-orchestratorService/config/utils"
	ceslogger "github.com/soluixdeveloper/ces-utilities/v2/ceslogger"
	cesresponse "github.com/soluixdeveloper/ces-utilities/v2/cesresponse"
)

func CreateWorkflow(metadata model.WorkflowMetadata) cesresponse.Response {
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
	err := dal.CreateWorkflow(&requestData)
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

func UpdateWorkflow(metadata model.WorkflowMetadata) cesresponse.Response {
	logging := ceslogger.Logger{}

	response := cesresponse.Response{
		StatusCode: utils.HTTPBadRequest,
	}

	// PREPARE REQUEST DATA
	requestData := metadata.Body
	requestData.ID = metadata.WorkflowQueryParams.ID
	requestData.UpdatedAt, _ = utils.TimestampNow(false, config.AppConfig.FormatTimestamp)

	// UPDATE DATABASE
	err := dal.UpdateWorkflow(&requestData)
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

func ListWorkflow(metadata model.WorkflowMetadata) cesresponse.Response {
	logging := ceslogger.Logger{}

	response := cesresponse.Response{
		StatusCode: utils.HTTPBadRequest,
	}

	Workflows := &[]model.Workflow{}
	var totalData int64

	queryResult := dal.ListWorkflow(metadata.WorkflowQueryParams, Workflows, &totalData)
	if queryResult != nil {
		logging.LogError("list customer type data from dal", queryResult.Error)
		errorMessage := helper.GenerateError(cesresponse.RCBadRequest, metadata.Header["locale"], nil)
		response.ResponseData = utils.GenerateJsonResponse(false, errorMessage, nil, cesresponse.RCBadRequest, nil)
		return response
	}

	response.StatusCode = utils.HTTPOk
	response.ResponseData = utils.GenerateJsonResponse(true, "", Workflows, nil, &totalData)
	return response
}

func GetWorkflow(metadata model.WorkflowMetadata) cesresponse.Response {
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

	var Workflow model.Workflow
	Workflow.ID = metaID

	queryResult := dal.GetWorkflow(&Workflow)
	if queryResult != nil {
		logging.LogError("get customer type data from dal", queryResult.Error)
		errorMessage := helper.GenerateError(cesresponse.RCDataNotFound, metadata.Header["locale"], nil)
		response.ResponseData = utils.GenerateJsonResponse(false, errorMessage, nil, cesresponse.RCDataNotFound, nil)
		return response
	}

	response.StatusCode = utils.HTTPOk
	response.ResponseData = utils.GenerateJsonResponse(true, "", Workflow, nil, nil)
	return response
}
