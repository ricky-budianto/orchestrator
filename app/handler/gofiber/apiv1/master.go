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

func MasterRouter(app fiber.Router) {
	app.Get(utils.PathID, MasterHandler)
	app.Get(utils.PathBase, MasterHandler)
	app.Post(utils.PathBase, MasterHandler)
	app.Put(utils.PathBase, MasterHandler)
}

func MasterHandler(c *fiber.Ctx) error {
	metadata := model.MasterMetadata{}
	metadata.Path = c.Path()
	metadata.Header = c.GetReqHeaders()
	metadata.BodyRaw = c.Body()
	metadata.Params = c.AllParams()
	metadata.Context = c

	if err := c.QueryParser(&metadata.QueryParams); err != nil {
		return c.Status(utils.HTTPBadRequest).JSON(utils.GenerateJsonResponse(false, "", nil, utils.RCBadRequest, nil))
	}

	if err := c.QueryParser(&metadata.MasterQueryParams); err != nil {
		return c.Status(utils.HTTPBadRequest).JSON(utils.GenerateJsonResponse(false, "", nil, utils.RCBadRequest, nil))
	}

	var response cesresponse.Response
	// log.Println("ini path", c.Route().Path)
	// log.Println("ini path2", c.Path())
	resourcePath := strings.TrimPrefix(c.Route().Path, utils.PathService+utils.PathV1+utils.PathMaster)
	// log.Println("resource", resourcePath)
	switch resourcePath + c.Method() {
	case utils.PathBase + "POST":
		if c.Body() != nil {
			if err := helper.BodyParseAndValidateStruct(c, &metadata.Body); err != nil {
				return c.Status(utils.HTTPBadRequest).JSON(utils.GenerateJsonResponse(false, "", nil, cesresponse.RCBadRequest, nil))
			}
		}

		response = servicev1.CreateMaster(metadata)
	case utils.PathBase + "PUT":
		if c.Body() != nil {
			if err := helper.BodyParseAndValidateStruct(c, &metadata.Body); err != nil {
				return c.Status(utils.HTTPBadRequest).JSON(utils.GenerateJsonResponse(false, "", nil, cesresponse.RCBadRequest, nil))
			}
		}

		response = servicev1.UpdateMaster(metadata)

	case utils.PathBase + "GET":

		response = servicev1.ListMaster(metadata)
	case utils.PathID + "GET":

		response = servicev1.GetMaster(metadata)
	}

	return c.Status(response.StatusCode).JSON(response.ResponseData)
}
