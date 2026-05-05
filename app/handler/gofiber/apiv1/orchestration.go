package apiv1

import (
	"encoding/json"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/soluixdeveloper/ces-orchestratorService/app/helper"
	"github.com/soluixdeveloper/ces-orchestratorService/app/model"
	"github.com/soluixdeveloper/ces-orchestratorService/app/service/servicev1"
	"github.com/soluixdeveloper/ces-orchestratorService/config"
	"github.com/soluixdeveloper/ces-utilities/v2/ceslogger"
	cesresponse "github.com/soluixdeveloper/ces-utilities/v2/cesresponse"
)

func OrchestrationRouter(app fiber.Router) {
	for _, v := range config.Endpoints {
		switch v.Method {
		case "GET":
			app.Get(v.Path, OrchestrationHandler)
		case "POST":
			app.Post(v.Path, OrchestrationHandler)
		case "DELETE":
			app.Delete(v.Path, OrchestrationHandler)
		case "PUT":
			app.Put(v.Path, OrchestrationHandler)
		}
	}
}

func OrchestrationHandler(c *fiber.Ctx) error {
	metadata := model.OrchestrationMetadata{}
	metadata.Path = c.Path()
	metadata.Header = c.GetReqHeaders()
	metadata.Headers = c.GetReqHeaders()
	metadata.BodyRaw = c.Body()
	metadata.Params = c.AllParams()
	metadata.Context = c
	metadata.Authorization = metadata.Header["Authorization"]
	json.Unmarshal(c.Body(), &metadata.Body)

	correlationId := string(c.Response().Header.Peek("X-Request-Id"))
	if correlationId == "" {
		correlationId = helper.UUIDgenerator()
	}
	logging := ceslogger.NewLogger(correlationId)
	logging.LogInfo("X-Request-Id: ", correlationId)

	queryString := strings.Split(string(c.Request().URI().QueryString()), "&")
	metadata.QueryParams = make(map[string][]string)

	for _, q := range queryString {
		query := strings.Split(q, "=")

		if len(query) > 1 {
			qVal := strings.Replace(q, query[0]+"=", "", 1)
			metadata.QueryParams[query[0]] = append(metadata.QueryParams[query[0]], qVal)
		}
	}

	var response cesresponse.Response
	resourcePath := c.Route().Path
	metadata.RequestType = config.RequestTypes[c.Method()+"_"+resourcePath]

	response = servicev1.CreateOrchestration(c.UserContext(), metadata, correlationId)

	return c.Status(response.StatusCode).JSON(response.ResponseData)
}
