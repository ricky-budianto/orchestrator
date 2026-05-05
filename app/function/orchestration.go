package function

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	jsonExt "github.com/json-iterator/go"
	"github.com/soluixdeveloper/ces-orchestratorService/app/function/dal"
	"github.com/soluixdeveloper/ces-orchestratorService/app/helper"
	"github.com/soluixdeveloper/ces-orchestratorService/app/model"
	"github.com/soluixdeveloper/ces-orchestratorService/config"
	"github.com/soluixdeveloper/ces-orchestratorService/config/elasticlog"
	orchestrationConfig "github.com/soluixdeveloper/ces-orchestratorService/config/orchestration"
	cacheredis "github.com/soluixdeveloper/ces-orchestratorService/config/redis"
	"github.com/soluixdeveloper/ces-orchestratorService/config/telemetry"
	"github.com/soluixdeveloper/ces-orchestratorService/config/utils"
	"github.com/soluixdeveloper/ces-utilities/v2/cesapp"
	"github.com/soluixdeveloper/ces-utilities/v2/ceslogger"
	"github.com/soluixdeveloper/ces-utilities/v2/cesmodel"
	"github.com/soluixdeveloper/ces-utilities/v2/cesrabbitmq"
	"github.com/soluixdeveloper/ces-utilities/v2/cesresponse"
	"github.com/soluixdeveloper/ces-utilities/v2/cesutils"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"gopkg.in/yaml.v2"
)

var workflowStateLock sync.RWMutex

func WorkflowInit() {
	logging := ceslogger.Logger{}

	// Get list workflow configuration with status active
	workflows := []model.WorkflowConfiguration{}
	offset := 0
	var totalData int64
	totalData = 1
	for len(workflows) < int(totalData) {
		var list []model.WorkflowConfiguration
		list, totalData = listWorkflow(offset, logging)
		workflows = append(workflows, list...)
		offset = len(list)
	}

	// Collect request type
	for _, configuration := range workflows {
		decodedWorkflow, err := helper.ConvertConfiguration(configuration)
		if !config.AppConfig.RedisUse {
			if err != nil {
				logging.LogError("Decode workflow congifuration", configuration.ID, err.Error())
				continue
			}
			config.WofkflowConfigurations[configuration.ID] = decodedWorkflow
		} else {
			err := cacheredis.WriteCache(configuration.ID, "workflow_configurations", decodedWorkflow)
			if err != nil {
				logging.LogError("Create memory cache workflow congifuration", configuration.ID, err.Error())
				config.WofkflowConfigurations[configuration.ID] = decodedWorkflow
				continue
			}
		}
		config.RequestTypes[configuration.Method+"_"+configuration.Path] = configuration.ID
		config.Endpoints = append(config.Endpoints, model.Endpoints{Path: configuration.Path, Method: configuration.Method})
	}
	logging.LogInfo("Total list of workflow configuration", len(config.Endpoints))
	logging.LogInfo("List of workflow configuration", config.Endpoints)
}

func listWorkflow(offset int, logging ceslogger.Logger) ([]model.WorkflowConfiguration, int64) {
	filter := model.WorkflowConfigurationHTTPQueryParameter{
		HTTPQueryParameter: model.HTTPQueryParameter{
			Limit:  500,
			Offset: offset,
		},
	}
	listWorkflow := []model.WorkflowConfiguration{}
	var totalData int64
	err := dal.ListWorkflowConfiguration(context.Background(), filter, &listWorkflow, &totalData)
	if err != nil {
		logging.LogFatal("List workflow configuration", err.Error())
	}
	return listWorkflow, totalData
}

func Orchestrate(ctx context.Context, metadata model.OrchestrationMetadata, logging ceslogger.Logger) (response cesresponse.Response) {
	if ctx == nil {
		ctx = context.Background()
	}
	var span trace.Span
	startTime := time.Now()

	// Initialize response
	response = cesresponse.Response{
		StatusCode: utils.HTTPBadRequest,
	}

	// Initialize telemetry if enabled
	if config.TelemetryConfig.TelemetryEnabled && telemetry.Tracer != nil {
		ctx, span = telemetry.Tracer.Start(ctx, "workflow.orchestrate",
			trace.WithAttributes(
				attribute.String("workflow.request_type", metadata.RequestType),
				attribute.String("workflow.correlation_id", metadata.CorrelationID),
			),
		)
		defer func() {
			// Record workflow metrics
			duration := time.Since(startTime).Seconds()

			// Derive status from status code
			status := "success"
			if response.StatusCode >= 400 {
				status = "error"
			}

			workflowLabels := []attribute.KeyValue{
				attribute.String("request_type", metadata.RequestType),
				attribute.String("status", status),
				attribute.Int("status_code", response.StatusCode),
			}

			if telemetry.WorkflowCounter != nil {
				telemetry.WorkflowCounter.Add(ctx, 1, metric.WithAttributes(workflowLabels...))
			}

			if telemetry.WorkflowDuration != nil {
				telemetry.WorkflowDuration.Record(ctx, duration, metric.WithAttributes(workflowLabels...))
			}

			// Set span attributes
			span.SetAttributes(
				attribute.String("workflow.status", status),
				attribute.Int("workflow.status_code", response.StatusCode),
				attribute.Float64("workflow.duration", duration),
			)

			// Set span status based on response
			if response.StatusCode >= 400 {
				span.SetStatus(codes.Error, "Workflow failed")
			}

			span.End()
		}()
	}
	var resData interface{}
	var globalVariable = map[string]interface{}{}
	resDataMap := model.ResponseData{}
	response.ResponseData = resData
	logging.LogInfo("START WORKFLOW ", metadata.RequestType)
	metadataBytes, _ := jsonExt.Marshal(metadata)
	logging.LogDebug("WORKFLOW DATA ", cesutils.MaskJSONForLog(metadataBytes, config.AppConfig.LogMaskDepth, config.AppConfig.LogMaskThreshold))
	if metadata.RequestType == "" {
		resDataMap.Message = "Invalid Request Type"
		resDataMap.ResponseCode = "E1"
		response.ResponseData = resDataMap
		response.StatusCode = 400
		return response
	}

	dataCenter := map[string]interface{}{}
	dataCenter["startEvent"] = metadata
	workload := model.YamlWorkload{}
	if config.AppConfig.RedisUse {
		// Get workflow configuration from redis
		getCaching, err := cacheredis.ReadCache(metadata.RequestType, "workflow_configurations")
		if err != nil {
			logging.LogError("ReadCache", err)
			workflowConfiguration := model.WorkflowConfiguration{
				ID: metadata.RequestType,
			}
			err = dal.GetWorkflowConfiguration(ctx, &workflowConfiguration)
			if err != nil {
				logging.LogDebug("Request types not found", metadata.RequestType)
				resDataMap.Message = "Invalid Workload File Orchestrator"
				resDataMap.ResponseCode = "E11"
				response.ResponseData = resDataMap
				response.StatusCode = 400
				return response
			}
			workload, err = helper.ConvertConfiguration(workflowConfiguration)
			if err != nil {
				logging.LogError("Decode workflow congifuration", workflowConfiguration.ID, err.Error())
				resDataMap.Message = "Invalid Workload File Orchestrator"
				resDataMap.ResponseCode = "E11"
				response.ResponseData = resDataMap
				response.StatusCode = 400
				return response
			}
			go func() {
				config.WofkflowConfigurations[metadata.RequestType] = workload
				cacheredis.WriteCache(workflowConfiguration.ID, "workflow_configurations", workload)
			}()
		} else {
			jsonExt.Unmarshal(getCaching, &workload)
			logging.LogDebug("Workflow Configuration from Cache", workload)
		}
	} else {
		if reflect.ValueOf(config.WofkflowConfigurations[metadata.RequestType]).IsZero() {
			logging.LogDebug("Request types not found", metadata.RequestType)
			resDataMap.Message = "Invalid Workload File Orchestrator"
			resDataMap.ResponseCode = "E11"
			response.ResponseData = resDataMap
			response.StatusCode = 400
			return response
		}
		workload = config.WofkflowConfigurations[metadata.RequestType]
	}
	startEventByte, _ := json.Marshal(workload.StartEvent)
	startEvent := model.Service{}
	json.Unmarshal(startEventByte, &startEvent)
	if targetRef, ok := workload.StartEvent.(map[interface{}]interface{}); ok {
		startEvent.TargetRef = append(startEvent.TargetRef, targetRef["targetRef"])
	} else if targetRef, ok := workload.StartEvent.(map[string]interface{}); ok {
		startEvent.TargetRef = append(startEvent.TargetRef, targetRef["targetRef"])

	}
	if workload.UniqueConstraint != nil {
		constraintResponse := helper.CollectData(workload.UniqueConstraint.Key, dataCenter, logging)
		logging.LogDebug("unique constraint validation", workload.UniqueConstraint, constraintResponse)
		type uniqueConstraint struct {
			Data    *string `json:"data"`
			Success bool    `json:"success"`
		}
		dataByte, _ := jsonExt.Marshal(constraintResponse)
		uniqueConstraintData := uniqueConstraint{}
		jsonExt.Unmarshal(dataByte, &uniqueConstraintData)
		logging.LogDebug("uniqueConstraint", uniqueConstraintData.Data)
		ctx := context.WithValue(context.Background(), utils.CorrelationIDKey, metadata.CorrelationID)
		if uniqueConstraintData.Data != nil {
			success, err := cacheredis.WriteCacheNotExists(ctx,
				*uniqueConstraintData.Data,
				config.AppConfig.ProjectModule,
				metadata.RequestType,
				time.Duration(workload.UniqueConstraint.TimeToLive)*1000000000,
			)
			if !success || err != nil {
				if err != nil {
					logging.LogError("unique constraint exceeded", *uniqueConstraintData.Data, err.Error())
				} else {
					err = errors.New("unique constraint exceeded")
				}
				response.ResponseData = utils.GenerateJsonResponse(false, err.Error(), *uniqueConstraintData.Data, "1050", nil)
				return response
			}
			deleteCache := func() {
				go cacheredis.DeleteCache(*uniqueConstraintData.Data,
					config.AppConfig.ProjectModule)
			}
			defer deleteCache()
		}
	}
	node := startEvent.TargetRef

	done := make(chan bool)

	// Create Workflow History
	var workflowHistory = model.WorkflowHistory{}
	workflowHistory.WorkflowConfigurationId = metadata.RequestType
	workflowHistory.Id = metadata.CorrelationID
	var wgWorkflowHistories sync.WaitGroup
	wgWorkflowHistories.Add(1)
	cesapp.Go(func() {
		defer wgWorkflowHistories.Done()
		metadataForHistory := metadata
		metadataForHistory.Authorization = maskAuthValue(metadata.Authorization)
		if metadata.Header != nil {
			headerCopy := make(map[string]string, len(metadata.Header))
			for k, v := range metadata.Header {
				if strings.EqualFold(k, "authorization") {
					headerCopy[k] = maskAuthValue(v)
				} else {
					headerCopy[k] = v
				}
			}
			metadataForHistory.Header = headerCopy
		}
		if metadata.Headers != nil {
			headersCopy := make(map[string]string, len(metadata.Headers))
			for k, v := range metadata.Headers {
				if strings.EqualFold(k, "authorization") {
					headersCopy[k] = maskAuthValue(v)
				} else {
					headersCopy[k] = v
				}
			}
			metadataForHistory.Headers = headersCopy
		}
		workflowHistory.Request = metadataForHistory
		workflowHistory.Status = "ACTIVE"
		if metadata.Header["X-User-Id"] != "" {
			workflowHistory.CreatedByID = metadata.Header["X-User-Id"]
		} else if metadata.Header["X-Customer-User-Id"] != "" {
			workflowHistory.CreatedByID = metadata.Header["X-Customer-User-Id"]
		} else {
			workflowHistory.CreatedByID = "SYSTEM"
		}
		workflowHistory.CreatedAt, _ = utils.TimestampNow(false, config.AppConfig.FormatTimestamp)
		workflowHistory.UpdateByID = workflowHistory.CreatedByID
		workflowHistory.UpdateByName = workflowHistory.CreatedByName
		workflowHistory.UpdatedAt, _ = utils.TimestampNow(false, config.AppConfig.FormatTimestamp)
		workflowHistory.RevisionNumber = 0
		dal.CreateWorkflowHistory(ctx, &workflowHistory)
	})

	// Update workflow history
	updateHistory := func(response interface{}) {
		wgWorkflowHistories.Wait()
		workflowHistory.Id = metadata.CorrelationID
		workflowHistory.Status = "EXECUTED"
		workflowHistory.Response = response
		workflowHistory.UpdatedAt, _ = utils.TimestampNow(false, config.AppConfig.FormatTimestamp)
		workflowHistory.RevisionNumber = 1
		dal.UpdateWorkflowHistory(ctx, metadata.CorrelationID, &workflowHistory)
	}

	go func() {
		var task = true
		var i = 0
		var workFlowStateID string
		for task {

			//check type
			service := workload.ServiceTask

			serviceTask := model.Service{}
			serviceTaskMap := map[interface{}]interface{}{}
			serviceTaskByte, _ := yaml.Marshal(service[node[0].(string)])

			yaml.Unmarshal(serviceTaskByte, &serviceTaskMap)
			serviceTaskConvert := helper.Convert(serviceTaskMap)
			dataByte, _ := json.Marshal(serviceTaskConvert)
			json.Unmarshal(dataByte, &serviceTask)

			switch serviceTask.Tipe {
			case "worker":
				//send request
				logging.LogDebug("Worker request ", serviceTask.Worker)
				newTime := time.Now()
				logging.LogDebug("Worker sourceRef ", helper.LogJSON(serviceTask.SourceRef))

				response := WorkerRequest(ctx, serviceTask, dataCenter, node[0].(string), logging, workflowHistory, &workFlowStateID)

				responseBytes, _ := jsonExt.Marshal(response)
				logging.LogDebug("Worker response at ", time.Since(newTime), serviceTask.Worker, "=>", cesutils.MaskJSONForLog(responseBytes, config.AppConfig.LogMaskDepth, config.AppConfig.LogMaskThreshold))

				//read response
				dataCenter[node[0].(string)] = response
				if len(serviceTask.SaveGlobal) > 0 {
					temp := globalVariable
					for k, v := range serviceTask.SaveGlobal {
						dataSave := helper.CollectData(v, dataCenter, logging)
						if dataSave.Message != "Success" {
							continue
						}
						temp[k] = dataSave.Data
					}
					globalVariable = temp
				}
				if len(globalVariable) > 0 {
					dataCenter["global_variable"] = globalVariable
				}
				i++
				node = serviceTask.TargetRef
			case "conditional":

				_, target := ConditionalState(ctx, serviceTask, dataCenter, &globalVariable, logging, workflowHistory, node[0].(string), &workFlowStateID, workload.DebugMode)

				if len(globalVariable) > 0 {
					dataCenter["global_variable"] = globalVariable
				}

				node[0] = target
			case "endEvent":
				response := EndEvent(serviceTask, dataCenter, logging)
				dataByteRes, e := jsonExt.Marshal(response)
				if e != nil {
					logging.LogError("error marshal", e)
				}
				jsonExt.Unmarshal(dataByteRes, &resData)
				task = false
				logging.LogDebug("Workflow Finished at ", time.Since(startTime).Milliseconds(), " : ", metadata.CorrelationID)
			default:
				resDataMap.Message = "Type of service " + node[0].(string) + " is undefined"
				resDataMap.Data = serviceTask

				task = false
				response.StatusCode = 400
				response.ResponseData = resDataMap
			}
		}
		done <- true
	}()

	// validasi timeout
	timeout := 90
	if workload.Timeout > 0 {
		timeout = workload.Timeout
	}
	select {
	case <-time.After(time.Duration(timeout) * time.Second):
		logging.LogInfo("timeout:", metadata.CorrelationID)
		resDataMap.Success = false
		resDataMap.Message = "Connection Timeout"
		resDataMap.ResponseCode = "1006"
		resDataMap.ErrorCode = "1006"
		response.StatusCode = 500
		response.ResponseData = resDataMap
		cesapp.Go(func() { updateHistory(response) })
		return response
	case <-done:
	}

	response.StatusCode = utils.HTTPOk

	dataByte0, _ := jsonExt.Marshal(resData)
	jsonExt.Unmarshal(dataByte0, &resDataMap)
	response.ResponseData = resData
	if resDataMap.StatusCode != 0 {
		response.StatusCode = resDataMap.StatusCode
		resDataMap.StatusCode = 0
	}
	cesapp.Go(func() { updateHistory(response) })

	return response
}

func WorkerRequest(ctx context.Context, serviceTask model.Service, dataCenter map[string]interface{}, reqType string, logging ceslogger.Logger, workflowHistory model.WorkflowHistory, workflowStateID *string) (res interface{}) {
	// Create state
	newTime := time.Now()
	stepName := reqType // Preserve the original step name for state logging
	res1 := model.ResponseData{}
	res1.Message = "Success"
	res1.ResponseCode = "S1"
	var sendData = map[string]interface{}{}
	//Preparing SourceRef
	switch serviceTask.SourceRefParsing {
	case orchestrationConfig.SourceRefParsing.Http:
		metadata := new(model.GofiberMetadataRabbit)
		for k, v := range serviceTask.SourceRef.(map[string]interface{}) {
			switch strings.ToLower(k) {
			case strings.ToLower(orchestrationConfig.HttpRequest.Auth):
				authData := model.ResponseData{}
				authMap, ok := v.(map[string]interface{})
				if ok {
					for k, v := range authMap {
						typeAuth := strings.ToLower(k)
						switch typeAuth {
						case "bearer":
							if strings.Contains(v.(string), "${{") {
								authData = helper.CollectData(v, dataCenter, logging)
								authData.Data = "Bearer " + authData.Data.(string)
								if authData.Message != "Success" {
									return authData
								}
							}
						case "basic":
							basicMap, ok := v.(map[string]interface{})
							if ok {
								if basicMap["username"] != nil && basicMap["password"] != nil {
									user, ok := basicMap["username"].(string)
									if !ok {
										authData.Message = "Failed : Encode Basic Auth {username}"
										authData.ResponseCode = "E0"
										return authData

									}
									pass, ok := basicMap["password"].(string)
									if !ok {
										authData.Message = "Failed : Encode Basic Auth {password}"
										authData.ResponseCode = "E0"
										return authData

									}
									authData.Data = "Basic " + base64.RawStdEncoding.EncodeToString([]byte(user+":"+pass))
								}
							} else {
								if strings.Contains(v.(string), "${{") {
									authData = helper.CollectData(v, dataCenter, logging)
									if authData.Message != "Success" {
										return authData
									}
								}
							}
						}
					}
				} else {
					authData = helper.CollectData(v, dataCenter, logging)
					if authData.Message != "Success" {
						return authData
					}
				}
				metadata.Header = make(map[string]string)
				metadata.Header[strings.Trim(orchestrationConfig.HttpRequest.Auth, "${}")] = authData.Data.(string)

				metadata.Headers = make(map[string]string)
				metadata.Headers[strings.Trim(orchestrationConfig.HttpRequest.Auth, "${}")] = authData.Data.(string)

			case strings.ToLower(orchestrationConfig.HttpRequest.Body):
				dataBody := helper.CollectData(v, dataCenter, logging)
				if dataBody.Message != "Success" {
					return dataBody
				}
				metadata.Body = dataBody.Data
				metadata.BodyRaw, _ = jsonExt.Marshal(dataBody.Data)

			case strings.ToLower(orchestrationConfig.HttpRequest.Params):
				dataParams := helper.CollectData(v, dataCenter, logging)
				if dataParams.Message != "Success" {
					return dataParams
				}

				params := map[string]string{}
				dataByte, _ := jsonExt.Marshal(dataParams.Data)
				jsonExt.Unmarshal(dataByte, &params)

				if params != nil {
					metadata.Params = params
				}

			case strings.ToLower(orchestrationConfig.HttpRequest.FormFile):
				dataFormFile := helper.CollectData(v, dataCenter, logging)
				if dataFormFile.Message != "Success" {
					return dataFormFile
				}
				metadata.FormFile[strings.Trim(orchestrationConfig.HttpRequest.FormFile, "${}")] = dataFormFile.Data

			case strings.ToLower(orchestrationConfig.HttpRequest.FormValue):
				dataFormValue := helper.CollectData(v, dataCenter, logging)
				if dataFormValue.Message != "Success" {
					return dataFormValue
				}
				metadata.FormValue[strings.Trim(orchestrationConfig.HttpRequest.FormValue, "${}")] = dataFormValue.Data

			case strings.ToLower(orchestrationConfig.HttpRequest.Query):
				dataQuery := helper.CollectData(v, dataCenter, logging)
				if dataQuery.Message != "Success" {
					return dataQuery
				}
				byteDataQuery, _ := jsonExt.Marshal(dataQuery.Data)
				jsonExt.Unmarshal(byteDataQuery, &metadata.QueryParams)
			case strings.ToLower(orchestrationConfig.HttpRequest.Headers), strings.ToLower(orchestrationConfig.HttpRequest.Header):
				dataHeader := helper.CollectData(v, dataCenter, logging)
				if dataHeader.Message != "Success" {
					return dataHeader
				}
				headerByte, _ := jsonExt.Marshal(dataHeader.Data)
				jsonExt.Unmarshal(headerByte, &metadata.Header)
				jsonExt.Unmarshal(headerByte, &metadata.Headers)
			default:
			}
		}
		if metadata.Header == nil {
			metadata.Header = make(map[string]string)
		}
		metadata.Header["X-Request-Id"] = workflowHistory.Id
		metadataByte, err := jsonExt.Marshal(metadata)
		if err != nil {
			logging.LogError("error marshal", err)
		}
		jsonExt.Unmarshal(metadataByte, &sendData)

	case orchestrationConfig.SourceRefParsing.Obj:
		mapData := helper.CollectData(serviceTask.SourceRef, dataCenter, logging)
		if mapData.Message != "Success" {
			return mapData
		}
		dataByte, _ := json.Marshal(mapData.Data)
		json.Unmarshal(dataByte, &sendData)
	}

	if serviceTask.RequestType != "" {
		reqType = serviceTask.RequestType
	}
	sendDataBytes, _ := jsonExt.Marshal(sendData)
	logging.LogDebug("Send Data to Worker ", reqType, " ; ", cesutils.MaskJSONForLog(sendDataBytes, config.AppConfig.LogMaskDepth, config.AppConfig.LogMaskThreshold))

	if serviceTask.Delay > 0 {
		logging.LogDebug("Delay process to Worker for a seconds", serviceTask.Delay)
		time.Sleep(time.Duration(serviceTask.Delay) * time.Second)
	}

	if serviceTask.Worker == "orchestrator_rpc" {
		request := model.OrchestrationMetadata{}
		dataByte, _ := json.Marshal(sendData)
		json.Unmarshal(dataByte, &request)
		request.RequestType = reqType
		request.Body = request.BodyRaw
		// res = WorkloadV3(request, ctx).ResponseData
		request.CorrelationID = helper.UUIDgenerator()
		logging := ceslogger.NewLogger(request.CorrelationID)

		// Create a child span for subworkflow execution
		subCtx := ctx
		var subSpan trace.Span
		if config.TelemetryConfig.TelemetryEnabled && telemetry.Tracer != nil {
			subCtx, subSpan = telemetry.Tracer.Start(ctx, "subworkflow."+reqType,
				trace.WithAttributes(
					attribute.String("subworkflow.request_type", reqType),
					attribute.String("subworkflow.correlation_id", request.CorrelationID),
					attribute.String("parent.correlation_id", workflowHistory.Id),
				),
			)
			defer subSpan.End()
		}

		response := Orchestrate(subCtx, request, *logging)
		responseByte, _ := jsonExt.Marshal(response.ResponseData)
		jsonExt.Unmarshal(responseByte, &res1)
		res1.ResponseData = response.ResponseData

		// Record subworkflow span status
		if config.TelemetryConfig.TelemetryEnabled && subSpan != nil {
			subSpan.SetAttributes(
				attribute.Int("subworkflow.status_code", response.StatusCode),
			)
			if response.StatusCode >= 400 {
				subSpan.SetStatus(codes.Error, "Subworkflow failed")
			}
		}
	} else {
		reqData := cesmodel.MessageQueue{
			Headers:       sendData["Header"],
			RequestType:   reqType,
			CorrelationID: helper.UUIDgenerator(),
			Value:         sendData,
		}
		if serviceTask.Timeout > 0 {
			reqData.Timeout = time.Duration(serviceTask.Timeout) * time.Second
		}
		if strings.Contains(serviceTask.Worker, "_rpc") {
			serviceTask.Worker = strings.ReplaceAll(serviceTask.Worker, "_rpc", "")
		}

		// Create a span for external worker call
		workerCtx := ctx
		var workerSpan trace.Span
		if config.TelemetryConfig.TelemetryEnabled && telemetry.Tracer != nil {
			workerCtx, workerSpan = telemetry.Tracer.Start(ctx, "worker."+serviceTask.Worker,
				trace.WithAttributes(
					attribute.String("worker.name", serviceTask.Worker),
					attribute.String("worker.request_type", reqType),
					attribute.String("worker.correlation_id", reqData.CorrelationID),
					attribute.String("workflow.id", workflowHistory.Id),
				),
			)
			defer workerSpan.End()
		}

		// Inject trace context into RabbitMQ headers for distributed tracing
		helper.InjectTraceContext(workerCtx, &reqData)

		response := cesrabbitmq.SendRPC(reqData, serviceTask.Worker)
		responseByte, _ := jsonExt.Marshal(response.ResponseData)
		jsonExt.Unmarshal(responseByte, &res1)
		res1.ResponseData = response.ResponseData

		// Record worker span status
		if config.TelemetryConfig.TelemetryEnabled && workerSpan != nil {
			workerSpan.SetAttributes(
				attribute.Int("worker.status_code", response.StatusCode),
			)
			if response.StatusCode >= 400 {
				workerSpan.SetStatus(codes.Error, "Worker call failed")
			}
		}
	}
	res = res1

	cesapp.Go(func() {
		var request, response interface{} = sendData, res1
		logResponse := model.ResponseData{}
		dataByteResponse, _ := jsonExt.Marshal(response)
		json.Unmarshal(dataByteResponse, &logResponse)
		createdAt, _ := utils.TimestampNow(false, config.AppConfig.FormatTimestamp)
		if elasticlog.ElasticLog != nil {
			elasticlog.ElasticLog.Log(model.OrchestrationLog{
				WorkflowId:          workflowHistory.Id,
				WorkflowRequestType: workflowHistory.WorkflowConfigurationId,
				Worker:              serviceTask.Worker,
				RequestType:         reqType,
				Request:             sendData,
				Response:            res1,
				StartTimestamp:      newTime,
				EndTimestamp:        createdAt,
				Task:                "worker",
			}, "orchestration_logs")
		}
		if serviceTask.RequestType == "" {
			serviceTask.RequestType = reqType
		}
		requestData := ""
		responseData := ""
		if requestMap, ok := request.(map[string]interface{}); ok {
			if requestMap["BodyRaw"] != nil {
				requestMap["BodyRaw"] = ""
			}
			requestDataByte, _ := json.Marshal(requestMap)
			requestData = string(requestDataByte)
			if len(requestData) > 5000 {
				requestData = requestData[:5000]
			}
		}
		responseDataByte, _ := json.Marshal(response)
		responseData = string(responseDataByte)
		// if responseMap, ok := response.(map[string]interface{}); ok {
		// 	if responseMap["data"] != nil {
		// 		responseMap["data"] = ""
		// 	}
		// 	if len(responseData) > 5000 {
		// 		responseData = responseData[:5000]
		// 	}
		// }

		currentState := map[string]model.State{
			stepName: {
				RequestSent:      newTime,
				ResponseReceived: createdAt,
				Success:          logResponse.Success,
				Message:          logResponse.Message,
				AdditionalInfo: map[string]interface{}{
					"request_data":  requestData,
					"response_data": responseData,
				},
			},
		}
		// Lock the entire read-modify-write sequence to prevent race conditions
		workflowStateLock.Lock()
		defer workflowStateLock.Unlock()

		var workflowState = model.WorkflowState{}
		if workflowStateID == nil || *workflowStateID == "" {
			audit := model.Audit{
				CreatedAt:      createdAt,
				CreatedByID:    workflowHistory.CreatedByID,
				CreatedByName:  workflowHistory.CreatedByName,
				UpdatedAt:      createdAt,
				UpdateByID:     workflowHistory.CreatedByID,
				UpdateByName:   workflowHistory.UpdateByName,
				RevisionNumber: 0,
			}
			workflowState.WorkflowId = workflowHistory.Id
			workflowState.State = currentState
			workflowState.Audit = audit
			workflowState.WorkflowRequestType = workflowHistory.WorkflowConfigurationId
			dal.CreateWorkflowState(ctx, &workflowState)
			*workflowStateID = workflowState.Id

		} else {
			workflowState.Id = *workflowStateID
			dal.GetWorkflowState(ctx, &workflowState)
			updateState := helper.MergeObject(workflowState.State, currentState)
			workflowStateUpdate := model.WorkflowState{}
			workflowStateUpdate.Id = *workflowStateID
			workflowStateUpdate.State = updateState
			dal.UpdateWorkflowState(ctx, &workflowStateUpdate)
		}
	})

	return res
}

func ConditionalState(ctx context.Context, serviceTask model.Service, dataCenter map[string]interface{}, globalVar *map[string]interface{}, logging ceslogger.Logger, workflowHistory model.WorkflowHistory, stepName string, workflowStateID *string, debugMode bool) (res bool, nextTarget string) {
	if ctx == nil {
		ctx = context.Background()
	}
	condition := model.Condition{}
	dataByte, _ := json.Marshal(serviceTask.Condition)
	json.Unmarshal(dataByte, &condition)
	var variableCheck []interface{}
	newTime := time.Now()
	for _, v := range condition.VariableCheck {
		varCheck := helper.CollectData(v, dataCenter, logging)
		variableCheck = append(variableCheck, varCheck.Data)
	}

	for _, m := range condition.Value {
		valid := true
		if len(m.Case) < 1 {
			m.Case = append(m.Case, "Success")
		}
		for len(m.Case) < len(variableCheck) {
			m.Case = append(m.Case, m.Case[len(m.Case)-1])
		}
		for a, b := range m.Case {
			stringB, ok := b.(string)
			if ok {
				if strings.Contains(stringB, "${{ibridge.Operator") {
					// switch m
					operator := strings.Split(stringB, "(")
					var comparator interface{}
					if len(operator) > 1 {
						// stringSplit := strings.Split(operator[1], "(")
						operatorValue := strings.ReplaceAll(stringB, operator[0], "")
						compTemp := operatorValue[1 : len(operatorValue)-1]
						if strings.ContainsAny(compTemp, "${}") {
							compData := helper.CollectData(compTemp, dataCenter, logging)
							if compData.Message != "Success" {
								break
							} else {
								comparator = compData.Data
							}
						} else {
							comparator = compTemp
						}
					}
					switch operator[0] {
					case orchestrationConfig.Operator.LenMore:
						value_b, ok := comparator.(int64)
						if !ok {
							value_ins, ok := comparator.(string)
							if !ok {
								valid = false
								break
							}
							value_ins_b, _ := strconv.Atoi(value_ins)
							value_b = int64(value_ins_b)
						}
						switch t := variableCheck[a].(type) {
						case map[string]interface{}:
							if !(len(t) > int(value_b)) {
								valid = false
							}
						case map[interface{}]interface{}:
							if !(len(t) > int(value_b)) {
								valid = false
							}
						case map[int]interface{}:
							if !(len(t) > int(value_b)) {
								valid = false
							}
						case map[float64]interface{}:
							if !(len(t) > int(value_b)) {
								valid = false
							}
						case string:
							if !(len(t) > int(value_b)) {
								valid = false
							}
						case []string:
							if !(len(t) > int(value_b)) {
								valid = false
							}
						case []int:
							if !(len(t) > int(value_b)) {
								valid = false
							}
						case []float64:
							if !(len(t) > int(value_b)) {
								valid = false
							}
						case []interface{}:
							if !(len(t) > int(value_b)) {
								valid = false
							}
						case []map[string]interface{}:
							if !(len(t) > int(value_b)) {
								valid = false
							}
						case []map[interface{}]interface{}:
							if !(len(t) > int(value_b)) {
								valid = false
							}
						default:
							valid = false
						}
					case orchestrationConfig.Operator.LenLess:
						value_b, ok := comparator.(int64)
						if !ok {
							value_ins, ok := comparator.(string)
							if !ok {
								valid = false
								break
							}
							value_ins_b, _ := strconv.Atoi(value_ins)
							value_b = int64(value_ins_b)
						}
						if variableCheck[a] == nil {
							break
						}
						switch t := variableCheck[a].(type) {
						case map[string]interface{}:
							if !(len(t) < int(value_b)) {
								valid = false
							}
						case map[interface{}]interface{}:
							if !(len(t) < int(value_b)) {
								valid = false
							}
						case map[int]interface{}:
							if !(len(t) < int(value_b)) {
								valid = false
							}
						case map[float64]interface{}:
							if !(len(t) < int(value_b)) {
								valid = false
							}
						case string:
							if !(len(t) < int(value_b)) {
								valid = false
							}
						case []string:
							if !(len(t) < int(value_b)) {
								valid = false
							}
						case []int:
							if !(len(t) < int(value_b)) {
								valid = false
							}
						case []float64:
							if !(len(t) < int(value_b)) {
								valid = false
							}
						case []interface{}:
							if !(len(t) < int(value_b)) {
								valid = false
							}
						case []map[string]interface{}:
							if !(len(t) < int(value_b)) {
								valid = false
							}
						case []map[interface{}]interface{}:
							if !(len(t) < int(value_b)) {
								valid = false
							}
						default:
							valid = false
						}
					case orchestrationConfig.Operator.LenEqual:
						value_b, ok := comparator.(int64)
						if !ok {
							value_ins, ok := comparator.(string)
							if !ok {
								valid = false
								break
							}
							value_ins_b, _ := strconv.Atoi(value_ins)
							value_b = int64(value_ins_b)
						}
						switch t := variableCheck[a].(type) {
						case map[string]interface{}:
							if !(len(t) == int(value_b)) {
								valid = false
							}
						case map[interface{}]interface{}:
							if !(len(t) == int(value_b)) {
								valid = false
							}
						case map[int]interface{}:
							if !(len(t) == int(value_b)) {
								valid = false
							}
						case map[float64]interface{}:
							if !(len(t) == int(value_b)) {
								valid = false
							}
						case string:
							if !(len(t) == int(value_b)) {
								valid = false
							}
						case []string:
							if !(len(t) == int(value_b)) {
								valid = false
							}
						case []int:
							if !(len(t) == int(value_b)) {
								valid = false
							}
						case []float64:
							if !(len(t) == int(value_b)) {
								valid = false
							}
						case []interface{}:
							if !(len(t) == int(value_b)) {
								valid = false
							}
						case []map[string]interface{}:
							if !(len(t) == int(value_b)) {
								valid = false
							}
						case []map[interface{}]interface{}:
							if !(len(t) == int(value_b)) {
								valid = false
							}
						default:
							valid = false
						}
					case orchestrationConfig.Operator.Nil:
						if variableCheck[a] != nil {
							valid = false
						}
					case orchestrationConfig.Operator.NotNil:
						if variableCheck[a] == nil {
							valid = false
						}
					case orchestrationConfig.Operator.Equal:

						if !reflect.DeepEqual(variableCheck[a], comparator) {
							valid = false
						}
					case orchestrationConfig.Operator.NotEqual:
						if reflect.DeepEqual(variableCheck[a], comparator) {
							valid = false
						}
					case orchestrationConfig.Operator.MoreThan:
						// if value string -> convert to float
						if variableCheck[a] != nil && reflect.TypeOf(variableCheck[a]).Kind() == reflect.String {
							valueCheck, err := strconv.Atoi(variableCheck[a].(string))
							if err != nil {
								logging.LogError("error format value ", orchestrationConfig.Operator.MoreThan, variableCheck[a])
								logging.LogError("error format value ", orchestrationConfig.Operator.MoreThan, variableCheck[a])
							} else {
								variableCheck[a] = float64(valueCheck)
							}
						}
						value_a, ok := variableCheck[a].(float64)
						if !ok {
							valid = false
							break
						}
						value_b, ok := comparator.(float64)
						if !ok {
							value_ins, ok := comparator.(string)
							if !ok {
								valid = false
								break
							}
							value_ins_b, _ := strconv.Atoi(value_ins)
							value_b = float64(value_ins_b)
						}
						if !(value_a > value_b) {
							valid = false
						}
					case orchestrationConfig.Operator.MoreThanEqual:
						value_a, ok := variableCheck[a].(float64)
						if !ok {
							value_ins, ok := variableCheck[a].(string)
							if ok {
								value_ins_b, _ := strconv.Atoi(value_ins)
								value_a = float64(value_ins_b)
							} else {
								value_int, ok := comparator.(int)
								if !ok {
									valid = false
									break
								}
								value_a = float64(value_int)

							}
						}
						value_b, ok := comparator.(float64)
						if !ok {
							value_ins, ok := comparator.(string)
							if ok {
								value_ins_b, _ := strconv.Atoi(value_ins)
								value_b = float64(value_ins_b)
							} else {
								value_int, ok := comparator.(int)
								if !ok {
									valid = false
									break
								}
								value_b = float64(value_int)
							}
						}
						if !(value_a >= value_b) {
							valid = false
						}
					case orchestrationConfig.Operator.LessThan:
						value_a, ok := variableCheck[a].(float64)
						if !ok {
							value_ins, ok := variableCheck[a].(string)
							if ok {
								value_ins_b, _ := strconv.Atoi(value_ins)
								value_a = float64(value_ins_b)
							} else {
								value_int, ok := comparator.(int)
								if !ok {
									valid = false
									break
								}
								value_a = float64(value_int)

							}
						}
						value_b, ok := comparator.(float64)
						if !ok {
							value_ins, ok := comparator.(string)
							if ok {
								value_ins_b, _ := strconv.Atoi(value_ins)
								value_b = float64(value_ins_b)
							} else {
								value_int, ok := comparator.(int)
								if !ok {
									valid = false
									break
								}
								value_b = float64(value_int)
							}
						}
						if !(value_a < value_b) {
							valid = false
						}
					case orchestrationConfig.Operator.LessThanEqual:

						value_a, ok := variableCheck[a].(float64)
						if !ok {
							value_ins, ok := variableCheck[a].(string)
							if ok {
								value_ins_b, _ := strconv.Atoi(value_ins)
								value_a = float64(value_ins_b)
							} else {
								value_int, ok := comparator.(int)
								if !ok {
									valid = false
									break
								}
								value_a = float64(value_int)

							}
						}
						value_b, ok := comparator.(float64)
						if !ok {
							value_ins, ok := comparator.(string)
							if ok {
								value_ins_b, _ := strconv.Atoi(value_ins)
								value_b = float64(value_ins_b)
							} else {
								value_int, ok := comparator.(int)
								if !ok {
									valid = false
									break
								}
								value_b = float64(value_int)
							}
						}
						if !(value_a <= value_b) {
							valid = false
						}
					case orchestrationConfig.Operator.Like:
						value_a, ok := variableCheck[a].(string)
						if !ok {
							valid = false
							break
						}
						value_b, ok := comparator.(string)
						if !ok {
							valid = false
							break
						}
						if !(strings.Contains(value_a, value_b)) {
							valid = false
						}
					case orchestrationConfig.Operator.NotLike:
						value_a, ok := variableCheck[a].(string)
						if !ok {
							valid = false
							break
						}
						value_b, ok := comparator.(string)
						if !ok {
							valid = false
							break
						}
						if strings.Contains(value_a, value_b) {
							valid = false
						}
					case orchestrationConfig.Operator.In:
						value_a, ok := variableCheck[a].(string)
						if !ok {
							valid = false
							break
						}
						value_b, ok := comparator.([]string)
						if !ok {
							valid = false
							break
						}
						for _, v := range value_b {
							if value_a == v {
								break
							}
						}
					case orchestrationConfig.Operator.NotIn:
						value_a, ok := variableCheck[a].(string)
						if !ok {
							valid = false
							break
						}
						value_b, ok := comparator.([]string)
						if !ok {
							valid = false
							break
						}
						for _, v := range value_b {
							if value_a == v {
								valid = false
								break
							}
						}
					case orchestrationConfig.Operator.Regex:
						value_a, ok := variableCheck[a].(string)
						if !ok {
							valid = false
							break
						}
						value_b, ok := comparator.(string)
						if !ok {
							valid = false
							break
						}
						var regex, _ = regexp.Compile(value_b)

						if !regex.MatchString(value_a) {
							valid = false
						}
					case orchestrationConfig.Operator.NotRegex:
						value_a, ok := variableCheck[a].(string)
						if !ok {
							valid = false
							break
						}
						value_b, ok := comparator.(string)
						if !ok {
							valid = false
							break
						}
						var regex, _ = regexp.Compile(value_b)

						if regex.MatchString(value_a) {
							valid = false
						}
					case orchestrationConfig.Operator.Contains:
						value_b, ok := comparator.(string)
						if !ok {
							valid = false
							break
						}
						switch t := variableCheck[a].(type) {
						case map[string]interface{}:
							isMet := false
							for a := range t {
								if a == value_b {
									isMet = true
								}
							}
							if !isMet {
								valid = false
							}
						case []interface{}:
							isMet := false
							for _, b := range t {
								if reflect.TypeOf(b).Kind() != reflect.String {
									continue
								}
								if b == value_b {
									isMet = true
								}
							}
							if !isMet {
								valid = false
							}
						case []string:
							isMet := false
							for _, b := range t {
								if b == value_b {
									isMet = true
								}
							}
							if !isMet {
								valid = false
							}
						}
					case orchestrationConfig.Operator.ObjectType:
						value_b, ok := comparator.(string)
						if !ok {
							valid = false
							break
						}
						objComparator := orchestrationConfig.ObjectOperator[value_b]
						if reflect.TypeOf(variableCheck[a]).Kind() != objComparator {
							valid = false
						}
					case orchestrationConfig.Operator.Any:
						// Always valid, no check needed
					}
				} else {
					if variableCheck[a] != b {
						valid = false
					}
				}
			} else {
				if variableCheck[a] != b {
					valid = false
				}
			}

		}
		if valid {
			if len(m.SaveGlobal) > 0 {
				temp := *globalVar
				for k, v := range m.SaveGlobal {
					dataSave := helper.CollectData(v, dataCenter, logging)
					if dataSave.Message != "Success" {
						continue
					}
					temp[k] = dataSave.Data
				}
				if m.StatusCode > 0 {
					temp["status_code"] = m.StatusCode
				}
				globalVar = &temp
			}

			target := m.Target
			varCheck := variableCheck
			cesapp.Go(func() {
				createdAt, _ := utils.TimestampNow(false, config.AppConfig.FormatTimestamp)

				// Only log conditional state to workflow_states if debugMode is enabled
				if debugMode {
					// Create current state for conditional step
					currentState := map[string]model.State{
						stepName: {
							RequestSent:      newTime,
							ResponseReceived: createdAt,
							Success:          true,
							Message:          "Condition matched",
							AdditionalInfo: map[string]interface{}{
								"variable_check": varCheck,
								"target":         target,
							},
						},
					}

					// Lock the entire read-modify-write sequence to prevent race conditions
					workflowStateLock.Lock()
					defer workflowStateLock.Unlock()

					var workflowState = model.WorkflowState{}
					if workflowStateID == nil || *workflowStateID == "" {
						audit := model.Audit{
							CreatedAt:      createdAt,
							CreatedByID:    workflowHistory.CreatedByID,
							CreatedByName:  workflowHistory.CreatedByName,
							UpdatedAt:      createdAt,
							UpdateByID:     workflowHistory.CreatedByID,
							UpdateByName:   workflowHistory.UpdateByName,
							RevisionNumber: 0,
						}
						workflowState.WorkflowId = workflowHistory.Id
						workflowState.State = currentState
						workflowState.Audit = audit
						workflowState.WorkflowRequestType = workflowHistory.WorkflowConfigurationId
						dal.CreateWorkflowState(ctx, &workflowState)
						*workflowStateID = workflowState.Id
					} else {
						workflowState.Id = *workflowStateID
						dal.GetWorkflowState(ctx, &workflowState)
						updateState := helper.MergeObject(workflowState.State, currentState)
						workflowStateUpdate := model.WorkflowState{}
						workflowStateUpdate.Id = *workflowStateID
						workflowStateUpdate.State = updateState
						dal.UpdateWorkflowState(ctx, &workflowStateUpdate)
					}
				}

				if elasticlog.ElasticLog != nil {
					elasticlog.ElasticLog.Log(model.OrchestrationLog{
						WorkflowId:          workflowHistory.Id,
						WorkflowRequestType: workflowHistory.WorkflowConfigurationId,
						VariableCheck:       varCheck,
						Task:                "conditional",
						TargetState:         target,
						StartTimestamp:      newTime,
						EndTimestamp:        createdAt,
					}, "orchestration_logs")
				}
			})

			return valid, m.Target
		}
	}
	temp := *globalVar
	if v, ok := temp["status_code"]; !ok || v == nil {
		temp["status_code"] = serviceTask.DefaultStatusCode
	}

	cesapp.Go(func() {
		varCheck := variableCheck
		createdAt, _ := utils.TimestampNow(false, config.AppConfig.FormatTimestamp)

		// Only log conditional state to workflow_states if debugMode is enabled
		if debugMode {
			// Create current state for conditional step (default case)
			currentState := map[string]model.State{
				stepName: {
					RequestSent:      newTime,
					ResponseReceived: createdAt,
					Success:          false,
					Message:          "No condition matched, using default",
					AdditionalInfo: map[string]interface{}{
						"variable_check": varCheck,
						"target":         serviceTask.DefaultTask,
					},
				},
			}

			// Lock the entire read-modify-write sequence to prevent race conditions
			workflowStateLock.Lock()
			defer workflowStateLock.Unlock()

			var workflowState = model.WorkflowState{}
			if workflowStateID == nil || *workflowStateID == "" {
				audit := model.Audit{
					CreatedAt:      createdAt,
					CreatedByID:    workflowHistory.CreatedByID,
					CreatedByName:  workflowHistory.CreatedByName,
					UpdatedAt:      createdAt,
					UpdateByID:     workflowHistory.CreatedByID,
					UpdateByName:   workflowHistory.UpdateByName,
					RevisionNumber: 0,
				}
				workflowState.WorkflowId = workflowHistory.Id
				workflowState.State = currentState
				workflowState.Audit = audit
				workflowState.WorkflowRequestType = workflowHistory.WorkflowConfigurationId
				dal.CreateWorkflowState(ctx, &workflowState)
				*workflowStateID = workflowState.Id
			} else {
				workflowState.Id = *workflowStateID
				dal.GetWorkflowState(ctx, &workflowState)
				updateState := helper.MergeObject(workflowState.State, currentState)
				workflowStateUpdate := model.WorkflowState{}
				workflowStateUpdate.Id = *workflowStateID
				workflowStateUpdate.State = updateState
				dal.UpdateWorkflowState(ctx, &workflowStateUpdate)
			}
		}

		if elasticlog.ElasticLog != nil {
			elasticlog.ElasticLog.Log(model.OrchestrationLog{
				WorkflowId:          workflowHistory.Id,
				WorkflowRequestType: workflowHistory.WorkflowConfigurationId,
				VariableCheck:       varCheck,
				Task:                "conditional",
				TargetState:         serviceTask.DefaultTask,
				StartTimestamp:      newTime,
				EndTimestamp:        createdAt,
			}, "orchestration_logs")
		}
	})

	return false, serviceTask.DefaultTask
}

func maskAuthValue(v string) string {
	if v == "" {
		return v
	}
	return "(" + strconv.Itoa(len(v)) + " chars)"
}

func EndEvent(serviceTask model.Service, dataCenter map[string]interface{}, logging ceslogger.Logger) (data interface{}) {
	resdata := model.ResponseData{}
	var endData model.ResponseData

	if serviceTask.SourceRef != nil {
		endData = helper.CollectData(serviceTask.SourceRef, dataCenter, logging)
		endDataBytes, _ := jsonExt.Marshal(endData)
		logging.LogDebug("endData ", cesutils.MaskJSONForLog(endDataBytes, config.AppConfig.LogMaskDepth, config.AppConfig.LogMaskThreshold))
	} else {
		endData.Success = true
	}
	i := 0
	for _, v := range dataCenter {
		if i == 0 {
			i++
			continue
		}
		dataByte, _ := json.Marshal(v)
		json.Unmarshal(dataByte, &resdata)
		resdata.Data = struct{}{}
		i++
	}
	saveGlobal, ok := dataCenter["global_variable"].(map[string]interface{})
	if ok {
		if saveGlobal["message"] != nil {
			endData.Message = saveGlobal["message"].(string)
		}
		if saveGlobal["error_code"] != nil {
			endData.ErrorCode = saveGlobal["error_code"].(string)
		}
		if saveGlobal["status_code"] != nil {
			switch v := saveGlobal["status_code"].(type) {
			case int:
				endData.StatusCode = v
			case float64:
				endData.StatusCode = int(v)
			}
		}
		if totalData, ok := saveGlobal["total_data"]; ok {
			var new64 int64
			switch v := totalData.(type) {
			case int:
				new64 = int64(v)
			case int64:
				endData.TotalData = &v
			case float64:
				new64 = int64(v)
			default:
				logging.LogDebug("Unsupported data type for total_data:", v)
				return
			}
			endData.TotalData = &new64
		}
		if endData.StatusCode >= 400 {
			endData.Success = false
			endData.ResponseCode = ""
		}
	}

	if endData.StatusCode <= 0 {
		endData.StatusCode = serviceTask.DefaultStatusCode
	}
	resdata.ResponseData = nil
	endDataByte, _ := jsonExt.Marshal(endData)
	jsonExt.Unmarshal(endDataByte, &resdata)
	return resdata
}
