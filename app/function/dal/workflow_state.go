package dal

import (
	"context"
	"encoding/json"
	"time"

	"github.com/soluixdeveloper/ces-orchestratorService/app/model"
	"github.com/soluixdeveloper/ces-orchestratorService/config"
	ceslogger "github.com/soluixdeveloper/ces-utilities/v2/ceslogger"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func CreateWorkflowState(ctx context.Context, data *model.WorkflowState) error {
	if ctx == nil {
		ctx = context.Background()
	}
	return config.PostgreDB.WithContext(ctx).Create(data).Error
}

func GetWorkflowState(ctx context.Context, data *model.WorkflowState) error {
	if ctx == nil {
		ctx = context.Background()
	}
	return config.PostgreDB.WithContext(ctx).First(data).Error
}

func ListWorkflowState(ctx context.Context, aqp model.WorkflowStateHTTPQueryParameter, data *[]model.WorkflowState, totalData *int64) error {
	if ctx == nil {
		ctx = context.Background()
	}
	logging := ceslogger.Logger{}

	// PREPARE FILTER DATA
	var filter model.WorkflowState
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

func UpdateWorkflowState(ctx context.Context, data *model.WorkflowState) error {
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

func ListCleanUpWorkflowState(ctx context.Context, data *[]string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	retentionDay := config.AppConfig.RETENTIONLOG
	if retentionDay < 1 {
		retentionDay = 7
	}
	sevenDaysAgo := time.Now().AddDate(0, 0, -retentionDay)
	return config.PostgreDB.WithContext(ctx).Model(&model.WorkflowState{}).
		Where("created_at < ?", sevenDaysAgo).
		Pluck("id", &data).Error
}

func DeleteWorkflowState(ctx context.Context, ids []string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	return config.PostgreDB.WithContext(ctx).Unscoped().Where("id IN ?", ids).Delete(&model.WorkflowState{}).Error
}
