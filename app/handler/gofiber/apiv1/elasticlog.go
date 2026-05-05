package apiv1

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/soluixdeveloper/ces-orchestratorService/app/model"
	"github.com/soluixdeveloper/ces-orchestratorService/app/service/servicev1"
	"github.com/soluixdeveloper/ces-orchestratorService/config/utils"
	cesresponse "github.com/soluixdeveloper/ces-utilities/v2/cesresponse"
)

func ElasticLogRouter(app fiber.Router) {
	app.Get(utils.PathBase, ElasticLogHandler)
}

func ElasticLogHandler(c *fiber.Ctx) error {
	metadata := model.ElasticLogMetadata{}
	metadata.Path = c.Path()
	metadata.Header = c.GetReqHeaders()
	metadata.BodyRaw = c.Body()
	metadata.Params = c.AllParams()
	metadata.Context = c

	if err := c.QueryParser(&metadata.ElasticLogHTTPQueryParameter); err != nil {
		return c.Status(utils.HTTPBadRequest).JSON(utils.GenerateJsonResponse(false, "", nil, utils.RCBadRequest, nil))
	}

	var response cesresponse.Response
	// log.Println("ini path", c.Route().Path)
	// log.Println("ini path2", c.Path())
	resourcePath := strings.TrimPrefix(c.Route().Path, utils.PathService+utils.PathV1+utils.PathElasticLog)
	// log.Println("resource", resourcePath)
	switch resourcePath + c.Method() {
	case utils.PathBase + "GET":

		response = servicev1.ListElasticLog(metadata)

	}

	return c.Status(response.StatusCode).JSON(response.ResponseData)
}
