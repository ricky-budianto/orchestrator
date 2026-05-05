package dal

import (
	"encoding/json"
	"errors"

	"github.com/soluixdeveloper/ces-orchestratorService/app/model"
	"github.com/soluixdeveloper/ces-orchestratorService/config"
	ceslogger "github.com/soluixdeveloper/ces-utilities/v2/ceslogger"
	cesresponse "github.com/soluixdeveloper/ces-utilities/v2/cesresponse"
)

func CreateWorkflow(data *model.Workflow) error {
	return config.PostgreDB.Create(data).Error
}

func GetWorkflow(data *model.Workflow) error {
	return config.PostgreDB.First(data).Error
}

func ListWorkflow(aqp model.WorkflowHTTPQueryParameter, data *[]model.Workflow, totalData *int64) error {
	logging := ceslogger.Logger{}

	// PREPARE FILTER DATA
	var filter model.Workflow
	aqpByte, _ := json.Marshal(aqp)
	err := json.Unmarshal(aqpByte, &filter)
	if err != nil {
		logging.LogError("err", err.Error())
	}

	// PREPARE QUERY
	queryResult := config.PostgreDB.Model(data).Where(filter).Count(totalData)

	if len(aqp.OrderBy) > 0 {
		if aqp.Desc == "true" {
			aqp.OrderBy += " desc"
		}
		queryResult = queryResult.Order(aqp.OrderBy)
	}

	if aqp.Offset != 0 {
		queryResult = queryResult.Offset(aqp.Offset)
	}

	if aqp.Limit != 0 {
		queryResult = queryResult.Limit(aqp.Limit)
	}

	queryResult.Find(data).Count(totalData)

	return queryResult.Error
}

func UpdateWorkflow(data *model.Workflow) error {
	logging := ceslogger.Logger{}
	tx := config.PostgreDB.Begin()

	// GET OLD DATA
	oldData := &model.Workflow{
		ID: data.ID,
	}
	dalGet := tx.First(oldData)
	if dalGet.Error != nil {
		logging.LogError("update Workflow ", dalGet.Error)
		tx.Rollback()
		return errors.New(cesresponse.RCDataNotFound)
	}

	err := tx.Updates(data)
	if err.Error != nil {
		logging.LogError("update Workflow ", err.Error)
		tx.Rollback()
		return errors.New(cesresponse.RCNoDataUpdated)
	}

	return tx.Commit().Error
}
