package servicev1

import (
	"context"

	"github.com/soluixdeveloper/ces-orchestratorService/app/function"
	"github.com/soluixdeveloper/ces-orchestratorService/app/model"
	ceslogger "github.com/soluixdeveloper/ces-utilities/v2/ceslogger"
	cesresponse "github.com/soluixdeveloper/ces-utilities/v2/cesresponse"
)

func CreateOrchestration(ctx context.Context, metadata model.OrchestrationMetadata, correlationId string) cesresponse.Response {
	logging := ceslogger.NewLogger(correlationId)

	metadata.CorrelationID = correlationId
	response := function.Orchestrate(ctx, metadata, *logging)

	return response
}
