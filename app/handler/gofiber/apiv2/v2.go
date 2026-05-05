package apiv2

import (
	"github.com/gofiber/fiber/v2"
	"github.com/soluixdeveloper/ces-orchestratorService/app/helper"
	"github.com/soluixdeveloper/ces-orchestratorService/config/utils"
)

func ApiV2(app fiber.Router) {
	v2 := app.Group(utils.PathV2).Use(helper.CreateActivity)
	OrchestrationRouter(v2.Group(utils.PathOrchestrate))
}
