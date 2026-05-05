package apiv1

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/soluixdeveloper/ces-orchestratorService/app/helper"
	"github.com/soluixdeveloper/ces-orchestratorService/app/model"
	"github.com/soluixdeveloper/ces-orchestratorService/app/service/servicev1"
	"github.com/soluixdeveloper/ces-orchestratorService/config/utils"
	cesresponse "github.com/soluixdeveloper/ces-utilities/v2/cesresponse"
)

func WorkflowHistoryRouter(app fiber.Router) {
	app.Get(utils.PathID, WorkflowHistoryHandler)
	app.Get(utils.PathBase, WorkflowHistoryHandler)
	app.Post(utils.PathBase, WorkflowHistoryHandler)
	app.Put(utils.PathBase, WorkflowHistoryHandler)
}

func WorkflowHistoryHandler(c *fiber.Ctx) error {
	metadata := model.WorkflowHistoryMetadata{}
	metadata.Path = c.Path()
	metadata.Header = c.GetReqHeaders()
	metadata.BodyRaw = c.Body()
	metadata.Params = c.AllParams()
	metadata.Context = c

	if err := c.QueryParser(&metadata.WorkflowHistoryQueryParams); err != nil {
		return c.Status(utils.HTTPBadRequest).JSON(utils.GenerateJsonResponse(false, "", nil, utils.RCBadRequest, nil))
	}

	var response cesresponse.Response
	// log.Println("ini path", c.Route().Path)
	// log.Println("ini path2", c.Path())
	resourcePath := strings.TrimPrefix(c.Route().Path, utils.PathService+utils.PathV1+utils.PathWorkflowHistory)
	// log.Println("resource", resourcePath)
	switch resourcePath + c.Method() {
	case utils.PathBase + "POST":
		if c.Body() != nil {
			if err := helper.BodyParseAndValidateStruct(c, &metadata.Body); err != nil {
				return c.Status(utils.HTTPBadRequest).JSON(utils.GenerateJsonResponse(false, "", nil, cesresponse.RCBadRequest, nil))
			}
		}

		response = servicev1.CreateWorkflowHistory(metadata)
	case utils.PathBase + "PUT":
		if c.Body() != nil {
			if err := helper.BodyParseAndValidateStruct(c, &metadata.Body); err != nil {
				return c.Status(utils.HTTPBadRequest).JSON(utils.GenerateJsonResponse(false, "", nil, cesresponse.RCBadRequest, nil))
			}
		}

		response = servicev1.UpdateWorkflowHistory(metadata)

	case utils.PathBase + "GET":

		response = servicev1.ListWorkflowHistory(metadata)
	case utils.PathID + "GET":

		response = servicev1.GetWorkflowHistory(metadata)
	}

	return c.Status(response.StatusCode).JSON(response.ResponseData)
}
