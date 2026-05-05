package helper

import (
	"encoding/json"

	"github.com/soluixdeveloper/ces-orchestratorService/app/function/dal"
	"github.com/soluixdeveloper/ces-orchestratorService/app/model"
	"github.com/soluixdeveloper/ces-orchestratorService/config"
	cacheredis "github.com/soluixdeveloper/ces-orchestratorService/config/redis"
	"github.com/soluixdeveloper/ces-orchestratorService/config/utils"
	ceslogger "github.com/soluixdeveloper/ces-utilities/v2/ceslogger"
)

var (
	moduleError  = "internal server error"
	messageError = "message is not define"
)

func GetErrorDesc(id string, locale string) (string, error) {
	logging := ceslogger.Logger{}

	if locale == "" {
		locale = config.AppConfig.LocalLanguage
		if locale == "" {
			locale = "ID"
		}
	}

	getCaching, err := cacheredis.ReadCache(id, moduleError)
	errData := new(model.Error)
	if err != nil {
		logging.LogError("ReadCache", err)
	}

	if getCaching == nil {
		errData.ID = id
		err := dal.GetError(errData)
		if err != nil {
			logging.LogError("queryResult", err)
			return "", err
		}

		listen := cacheredis.WriteCache(id, moduleError, errData)
		logging.LogInfo("listen", listen)
		return errData.Descriptions[locale], nil
	}

	json.Unmarshal(getCaching, errData)

	return errData.Descriptions[locale], nil
}

func GenerateError(id string, locale string, data interface{}) string {
	logging := ceslogger.Logger{}

	resMsg, err := GetErrorDesc(id, locale)
	if err != nil {
		logging.LogError("getErrorDesc", err)
		resMsg = messageError
	}

	resMsg, err = utils.ParseTemplateToString(resMsg, data)
	if err != nil {
		logging.LogError("utils.ParseTemplateToString", err)
	}

	return resMsg
}
