package helper

import (
	"encoding/json"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/soluixdeveloper/ces-orchestratorService/app/model"
	"github.com/soluixdeveloper/ces-orchestratorService/config/utils"
	"github.com/soluixdeveloper/ces-orchestratorService/config/validator"
	"github.com/soluixdeveloper/ces-utilities/v2/cesapp"
	ceslogger "github.com/soluixdeveloper/ces-utilities/v2/ceslogger"
	"github.com/soluixdeveloper/ces-utilities/v2/cesmodel"
	"github.com/soluixdeveloper/ces-utilities/v2/cesrabbitmq"
)

// BodyParseAndValidateStruct parse request body validates the input struct
func BodyParseAndValidateStruct(ctx *fiber.Ctx, payload interface{}) error {
	var err error
	// parse using gofiber ctx.BodyParse
	if err = ctx.BodyParser(payload); err != nil {
		return err
	}

	// validate struct
	if err = validator.StructValidator(payload); err != nil {
		return err
	}

	return err
}

func CreateActivity(c *fiber.Ctx) error {
	logging := ceslogger.NewLogger("")
	// Read Mandatory header
	headers := c.GetReqHeaders()
	if headerAuth, ok := headers["Authorization"]; !ok || headerAuth == "" {
		logging.LogInfo("Skip Create User Activity - No access token")
		return c.Next()
	} else if headerActModule, ok := headers["X-Activity-Module"]; !ok || headerActModule == "" {
		logging.LogInfo("Skip Create User Activity - No activity module")
		return c.Next()
	} else if headerActTopic, ok := headers["X-Activity-Topic"]; !ok || headerActTopic == "" {
		logging.LogInfo("Skip Create User Activity - No activity topic")
		return c.Next()
	} else if strings.Contains(c.Path(), "login") {
		logging.LogInfo("Skip Create User Activity - login endpoint")
		return c.Next()
	} else {
		// check if Token Empty Bearer
		if len(headerAuth) < 8 {
			return c.Next()
		}

		var data model.Activity
		data.CommunicationProtocol = "HTTP"

		// Eead Request Body
		var reqBody map[string]interface{}
		json.Unmarshal(c.Body(), &reqBody)
		var reqType, ok = reqBody["type"].(string)
		if ok && reqType != "" {
			data.ActivityDetails.RequestType = reqType
		}

		// Read Headers IP
		if headerIPFwd, ok := headers["X-Forwarded-For"]; ok && headerIPFwd != "" {
			data.SourceIPAddress = headerIPFwd
		} else if headerIPAdd, ok := headers["X-Forwarded-Ip"]; ok && headerIPAdd != "" {
			data.SourceIPAddress = headerIPAdd
		}

		// Define Path
		data.ActivityDetails.EndpointPath = c.Path()
		if headerEndpointName, ok := headers["X-Endpoint-Name"]; !ok || headerEndpointName == "" {
			data.Name = strings.ReplaceAll(data.ActivityDetails.EndpointPath, "/", " ")
		} else {
			data.Name = headerEndpointName
		}

		// logging.LogInfo("targetModule", headerActModule)
		// logging.LogInfo("targetTopic", headerActTopic)
		// logging.LogInfo("data", data)

		// produce event
		var sendRequestData model.GofiberMetadata
		sendRequestData.BodyRaw, _ = json.Marshal(data)
		sendRequestData.Header = headers

		requestRPC := cesmodel.MessageQueue{
			RequestType:   headerActTopic,
			CorrelationID: utils.UUIDgenerator(),
			Value:         sendRequestData,
		}

		// Inject trace context for distributed tracing
		InjectTraceContext(c.UserContext(), &requestRPC)

		// send RPC
		cesapp.Go(func() { cesrabbitmq.SendRPC(requestRPC, headerActModule) })

		// Continue
		return c.Next()
	}
}
