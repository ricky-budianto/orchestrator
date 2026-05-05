package servicev1

import (
	"context"
	"sync"

	"github.com/soluixdeveloper/ces-orchestratorService/app/function/dal"
	"github.com/soluixdeveloper/ces-orchestratorService/app/helper"
	"github.com/soluixdeveloper/ces-orchestratorService/app/model"
	"github.com/soluixdeveloper/ces-orchestratorService/config"

	cacheredis "github.com/soluixdeveloper/ces-orchestratorService/config/redis"
	"github.com/soluixdeveloper/ces-orchestratorService/config/utils"
	"github.com/soluixdeveloper/ces-utilities/v2/cesapp"

	ceslogger "github.com/soluixdeveloper/ces-utilities/v2/ceslogger"

	cesresponse "github.com/soluixdeveloper/ces-utilities/v2/cesresponse"
)

func CreateWorkflowConfiguration(ctx context.Context, metadata model.WorkflowConfigurationMetadata) cesresponse.Response {
	if ctx == nil {
		ctx = context.Background()
	}
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
	requestData.RevisionNumber = 0

	// CREATE DATA TO DATABASE
	err := dal.CreateWorkflowConfiguration(ctx, &requestData)
	if err != nil {
		logging.LogError("create data to database using dal", err)
		errorMessage := helper.GenerateError(err.Error(), metadata.Header["locale"], nil)
		response.ResponseData = utils.GenerateJsonResponse(false, errorMessage, nil, err.Error(), nil)
		return response
	}
	cesapp.Go(func() {
		// Create Workflow Audit
		workflowAudit := model.WorkflowAudit{
			ResourceId:    requestData.ID,
			Data:          requestData,
			CreatedAt:     requestData.CreatedAt,
			CreatedByID:   requestData.CreatedByID,
			CreatedByName: requestData.CreatedByName,
			UpdatedAt:     requestData.UpdatedAt,
			UpdateByID:    requestData.UpdateByID,
			UpdateByName:  requestData.UpdateByName,
		}
		dal.CreateWorkflowAudit(context.Background(), &workflowAudit)
	})
	// Create memory cache
	decodedWorkflow, err := helper.ConvertConfiguration(requestData)
	if config.AppConfig.RedisUse {
		err = cacheredis.WriteCache(requestData.ID, "workflow_configurations", decodedWorkflow)
		if err != nil {
			logging.LogError("Create memory cache workflow congifuration", requestData.ID, err.Error())
			config.WofkflowConfigurations[requestData.ID] = decodedWorkflow
		}
	} else {
		config.WofkflowConfigurations[requestData.ID] = decodedWorkflow
	}

	response.StatusCode = utils.HTTPOk
	response.ResponseData = utils.GenerateJsonResponse(true, "", requestData, "", nil)
	return response
}

func UpdateWorkflowConfiguration(ctx context.Context, metadata model.WorkflowConfigurationMetadata) cesresponse.Response {
	if ctx == nil {
		ctx = context.Background()
	}
	logging := ceslogger.Logger{}

	response := cesresponse.Response{
		StatusCode: utils.HTTPBadRequest,
	}

	// PREPARE REQUEST DATA
	requestData := metadata.Body
	requestData.UpdatedAt, _ = utils.TimestampNow(false, config.AppConfig.FormatTimestamp)

	// UPDATE DATABASE
	err := dal.UpdateWorkflowConfiguration(ctx, &requestData)
	if err != nil {
		logging.LogError("update data to database using dal", err)
		errorMessage := helper.GenerateError(err.Error(), metadata.Header["locale"], nil)
		response.ResponseData = utils.GenerateJsonResponse(false, errorMessage, nil, err.Error(), nil)
		return response
	}
	cesapp.Go(func() {
		// Create Workflow Audit
		workflowAudit := model.WorkflowAudit{
			ResourceId:     requestData.ID,
			Data:           requestData,
			RevisionNumber: requestData.RevisionNumber,
			CreatedAt:      requestData.CreatedAt,
			CreatedByID:    requestData.CreatedByID,
			CreatedByName:  requestData.CreatedByName,
			UpdatedAt:      requestData.UpdatedAt,
			UpdateByID:     requestData.UpdateByID,
			UpdateByName:   requestData.UpdateByName,
		}
		dal.CreateWorkflowAudit(context.Background(), &workflowAudit)
	})
	decodedWorkflow, err := helper.ConvertConfiguration(requestData)
	if config.AppConfig.RedisUse {
		err = cacheredis.DeleteCache(requestData.ID, "workflow_configurations")
		if err != nil {
			logging.LogError("Delete memory cache workflow congifuration", requestData.ID, err.Error())
			config.WofkflowConfigurations[requestData.ID] = decodedWorkflow
		}
		err = cacheredis.WriteCache(requestData.ID, "workflow_configurations", decodedWorkflow)
		if err != nil {
			logging.LogError("Create memory cache workflow congifuration", requestData.ID, err.Error())
			config.WofkflowConfigurations[requestData.ID] = decodedWorkflow
		}
	} else {
		config.WofkflowConfigurations[requestData.ID] = decodedWorkflow
	}
	response.StatusCode = utils.HTTPOk
	response.ResponseData = utils.GenerateJsonResponse(true, "", requestData, "", nil)
	return response
}

func ListWorkflowConfiguration(ctx context.Context, metadata model.WorkflowConfigurationMetadata) cesresponse.Response {
	if ctx == nil {
		ctx = context.Background()
	}
	logging := ceslogger.Logger{}

	response := cesresponse.Response{
		StatusCode: utils.HTTPBadRequest,
	}

	WorkflowConfigurations := &[]model.WorkflowConfiguration{}
	var totalData int64

	queryResult := dal.ListWorkflowConfiguration(ctx, metadata.WorkflowConfigurationQueryParams, WorkflowConfigurations, &totalData)
	if queryResult != nil {
		logging.LogError("list customer type data from dal", queryResult.Error)
		errorMessage := helper.GenerateError(cesresponse.RCBadRequest, metadata.Header["locale"], nil)
		response.ResponseData = utils.GenerateJsonResponse(false, errorMessage, nil, cesresponse.RCBadRequest, nil)
		return response
	}

	response.StatusCode = utils.HTTPOk
	response.ResponseData = utils.GenerateJsonResponse(true, "", WorkflowConfigurations, nil, &totalData)
	return response
}

func GetWorkflowConfiguration(ctx context.Context, metadata model.WorkflowConfigurationMetadata) cesresponse.Response {
	if ctx == nil {
		ctx = context.Background()
	}
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

	var WorkflowConfiguration model.WorkflowConfiguration
	WorkflowConfiguration.ID = metaID

	queryResult := dal.GetWorkflowConfiguration(ctx, &WorkflowConfiguration)
	if queryResult != nil {
		logging.LogError("get customer type data from dal", queryResult.Error)
		errorMessage := helper.GenerateError(cesresponse.RCDataNotFound, metadata.Header["locale"], nil)
		response.ResponseData = utils.GenerateJsonResponse(false, errorMessage, nil, cesresponse.RCDataNotFound, nil)
		return response
	}

	response.StatusCode = utils.HTTPOk
	response.ResponseData = utils.GenerateJsonResponse(true, "", WorkflowConfiguration, nil, nil)
	return response
}

func DeleteWorkflowConfiguration(ctx context.Context, metadata model.WorkflowConfigurationMetadata) cesresponse.Response {
	if ctx == nil {
		ctx = context.Background()
	}
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

	var WorkflowConfiguration model.WorkflowConfiguration
	WorkflowConfiguration.ID = metaID

	err := dal.DeleteWorkflowConfiguration(ctx, &WorkflowConfiguration)
	if err != nil {
		logging.LogError("delete workflow configuration data from dal", err.Error)
		errorMessage := helper.GenerateError(cesresponse.RCDataNotFound, metadata.Header["locale"], nil)
		response.ResponseData = utils.GenerateJsonResponse(false, errorMessage, nil, cesresponse.RCDataNotFound, nil)
		return response
	}
	if config.AppConfig.RedisUse {
		err = cacheredis.DeleteCache(metaID, "workflow_configurations")
		if err != nil {
			logging.LogError("Delete memory cache workflow congifuration", metaID, err.Error())
		}
	} else {
		var syncMutex *sync.RWMutex
		syncMutex.Lock()
		delete(config.WofkflowConfigurations, metaID)
		syncMutex.Unlock()
		// config.WofkflowConfigurations[metaID] = decodedWorkflow
	}
	response.StatusCode = utils.HTTPOk
	response.ResponseData = utils.GenerateJsonResponse(true, "", WorkflowConfiguration, nil, nil)
	return response
}
