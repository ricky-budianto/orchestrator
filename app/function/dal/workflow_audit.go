package dal

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/soluixdeveloper/ces-orchestratorService/app/model"
	"github.com/soluixdeveloper/ces-orchestratorService/config"
	ceslogger "github.com/soluixdeveloper/ces-utilities/v2/ceslogger"
	cesresponse "github.com/soluixdeveloper/ces-utilities/v2/cesresponse"
)

func CreateWorkflowAudit(ctx context.Context, data *model.WorkflowAudit) error {
	if ctx == nil {
		ctx = context.Background()
	}
	return config.PostgreDB.WithContext(ctx).Create(data).Error
}

func GetWorkflowAudit(ctx context.Context, data *model.WorkflowAudit) error {
	if ctx == nil {
		ctx = context.Background()
	}
	return config.PostgreDB.WithContext(ctx).First(data).Error
}

func ListWorkflowAudit(ctx context.Context, aqp model.WorkflowAuditHTTPQueryParameter, data *[]model.WorkflowAudit, totalData *int64) error {
	if ctx == nil {
		ctx = context.Background()
	}
	logging := ceslogger.Logger{}

	// PREPARE FILTER DATA
	var filter model.WorkflowAudit
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

func UpdateWorkflowAudit(ctx context.Context, data *model.WorkflowAudit) error {
	if ctx == nil {
		ctx = context.Background()
	}
	logging := ceslogger.Logger{}
	tx := config.PostgreDB.WithContext(ctx).Begin()

	// GET OLD DATA
	oldData := &model.WorkflowAudit{
		ID: data.ID,
	}
	dalGet := tx.First(oldData)
	if dalGet.Error != nil {
		logging.LogError("update WorkflowAudit ", dalGet.Error)
		tx.Rollback()
		return errors.New(cesresponse.RCDataNotFound)
	}

	err := tx.Updates(data)
	if err.Error != nil {
		logging.LogError("update WorkflowAudit ", err.Error)
		tx.Rollback()
		return errors.New(cesresponse.RCNoDataUpdated)
	}

	return tx.Commit().Error
}
