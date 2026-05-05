package dal

import (
	"context"
	"encoding/json"

	"github.com/soluixdeveloper/ces-orchestratorService/app/model"
	"github.com/soluixdeveloper/ces-orchestratorService/config"
	ceslogger "github.com/soluixdeveloper/ces-utilities/v2/ceslogger"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func CreateWorkflowConfiguration(ctx context.Context, data *model.WorkflowConfiguration) error {
	if ctx == nil {
		ctx = context.Background()
	}
	return config.PostgreDB.WithContext(ctx).Create(data).Error
}

func GetWorkflowConfiguration(ctx context.Context, data *model.WorkflowConfiguration) error {
	if ctx == nil {
		ctx = context.Background()
	}
	return config.PostgreDB.WithContext(ctx).First(data).Error
}

func DeleteWorkflowConfiguration(ctx context.Context, data *model.WorkflowConfiguration) error {
	if ctx == nil {
		ctx = context.Background()
	}
	return config.PostgreDB.WithContext(ctx).Delete(data).Error
}

func ListWorkflowConfiguration(ctx context.Context, aqp model.WorkflowConfigurationHTTPQueryParameter, data *[]model.WorkflowConfiguration, totalData *int64) error {
	if ctx == nil {
		ctx = context.Background()
	}
	logging := ceslogger.Logger{}

	// PREPARE FILTER DATA
	var filter model.WorkflowConfiguration
	aqpByte, _ := json.Marshal(aqp)
	err := json.Unmarshal(aqpByte, &filter)
	if err != nil {
		logging.LogError("err", err.Error())
	}

	// PREPARE QUERY
	queryResult := config.PostgreDB.WithContext(ctx).Model(data).Where(filter).Count(totalData)

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

func ListWorkflowConfigurationV2(ctx context.Context, aqp model.WorkflowConfigurationHTTPQueryParameter, data *[]model.WorkflowConfiguration, totalData *int64) error {
	if ctx == nil {
		ctx = context.Background()
	}
	logging := ceslogger.Logger{}

	// PREPARE FILTER DATA
	var filter model.WorkflowConfiguration
	aqpByte, _ := json.Marshal(aqp)
	err := json.Unmarshal(aqpByte, &filter)
	if err != nil {
		logging.LogError("err", err.Error())
	}

	// PREPARE QUERY
	queryResult := config.PostgreDB.WithContext(ctx).Model(data).Select("id", "path", "method", "name").Where(filter).Count(totalData)

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

func UpdateWorkflowConfiguration(ctx context.Context, data *model.WorkflowConfiguration) error {
	if ctx == nil {
		ctx = context.Background()
	}
	tx := config.PostgreDB.WithContext(ctx).Clauses(clause.Returning{}).Updates(data)
	if tx.Error != nil {
		return tx.Error
	} else if tx.RowsAffected <= 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
