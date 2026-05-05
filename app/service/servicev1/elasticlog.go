package servicev1

import (
	"github.com/soluixdeveloper/ces-orchestratorService/app/model"
	"github.com/soluixdeveloper/ces-orchestratorService/config/elasticlog"
	"github.com/soluixdeveloper/ces-orchestratorService/config/utils"
	ceslogger "github.com/soluixdeveloper/ces-utilities/v2/ceslogger"
	cesresponse "github.com/soluixdeveloper/ces-utilities/v2/cesresponse"
)

func ListElasticLog(metadata model.ElasticLogMetadata) cesresponse.Response {
	logging := ceslogger.Logger{}

	response := cesresponse.Response{
		StatusCode: utils.HTTPBadRequest,
	}

	if metadata.ElasticLogHTTPQueryParameter.Index == "" {
		logging.LogError("list elastic log with no index", metadata.QueryParams)
		response.ResponseData = utils.GenerateJsonResponse(false, "undefined index", nil, cesresponse.RCBadRequest, nil)
		return response
	}

	elasticLog, err := elasticlog.ElasticLog.Query(metadata.ElasticLogHTTPQueryParameter)
	if err != nil {
		logging.LogError("list elastic log", err.Error())
		response.ResponseData = utils.GenerateJsonResponse(false, err.Error(), nil, cesresponse.RCBadRequest, nil)
		return response
	}
	totalData := int64(len(elasticLog))

	response.StatusCode = utils.HTTPOk
	response.ResponseData = utils.GenerateJsonResponse(true, "", elasticLog, nil, &totalData)
	return response
}
