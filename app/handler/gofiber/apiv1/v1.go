package apiv1

import (
	"github.com/gofiber/fiber/v2"
	"github.com/soluixdeveloper/ces-orchestratorService/app/helper"
	"github.com/soluixdeveloper/ces-orchestratorService/config/utils"
)

func ApiV1(app fiber.Router) {
	v1 := app.Group(utils.PathV1).Use(helper.CreateActivity)

	WorkflowAuditRouter(v1.Group(utils.PathWorkflowAudit))
	WorkflowConfigurationRouter(v1.Group(utils.PathWorkflowConfiguration))
	WorkflowHistoryRouter(v1.Group(utils.PathWorkflowHistory))
	WorkflowStateRouter(v1.Group(utils.PathWorkflowState))
	ElasticLogRouter(v1.Group(utils.PathElasticLog))
}

func ApiV1Orchestration(app fiber.Router) {
	OrchestrationRouter(app.Use(helper.CreateActivity))
}
