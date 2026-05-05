package helper

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"strconv"

	jsonExt "github.com/json-iterator/go"
	"github.com/kevinburke/nacl"
	"github.com/soluixdeveloper/ces-orchestratorService/app/model"
	"github.com/soluixdeveloper/ces-orchestratorService/config"
	"github.com/soluixdeveloper/ces-orchestratorService/config/telemetry"
	cesmodel "github.com/soluixdeveloper/ces-utilities/v2/cesmodel"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"golang.org/x/crypto/nacl/box"
	"gopkg.in/yaml.v2"
)

// CalculateCheckDigit calculates next check digit from given string
func CalculateCheckDigit(s string) int {
	totalDigit := 0
	for idx, c := range s {
		parsed, _ := strconv.Atoi(fmt.Sprintf("%c", c))
		if (len(s)-idx)%2 == 0 {
			totalDigit += parsed
		} else {
			parsed *= 2
			if parsed > 9 {
				totalDigit += parsed - 9
			} else {
				totalDigit += parsed
			}
		}
	}
	return (10 - (totalDigit % 10)) % 10
}

func PrecisionFloatWithoutRounding(input float64, decimal int) float64 {
	decimalMultiplier := 1
	for i := 0; i < decimal; i++ {
		decimalMultiplier *= 10
	}
	return float64(int(input*float64(decimalMultiplier))) / float64(decimalMultiplier)
}

func RoundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}

func NaClEncryption(data interface{}, key string) (interface{}, error) {
	var err error
	var dataByte []byte
	switch v := data.(type) {
	case string:
		dataByte = []byte(v)
	case map[string]interface{}:
		dataByte, err = json.Marshal(v)
		if err != nil {
			return nil, err
		}
	}

	senderPublic, senderPrivate, errGenKey := box.GenerateKey(rand.Reader)
	if errGenKey != nil {
		return nil, errGenKey
	}

	nonce := nacl.NewNonce()
	recPub, _ := base64.StdEncoding.DecodeString(key)
	var pubkey [32]byte

	copy(pubkey[:], recPub)
	var adrnonce [24]byte
	var adrkeyOut [32]byte
	adrnonce = *nonce
	adrkeyOut = *senderPublic
	encData := box.Seal(nil, dataByte, nonce, &pubkey, senderPrivate)

	encodedData := base64.StdEncoding.EncodeToString(encData)

	hashedData := map[string]interface{}{
		"key":   base64.StdEncoding.EncodeToString(adrkeyOut[:]),
		"data":  encodedData,
		"nonce": base64.StdEncoding.EncodeToString(adrnonce[:]),
	}

	return hashedData, nil
}

func ConvertConfiguration(data model.WorkflowConfiguration) (model.YamlWorkload, error) {
	var res model.YamlWorkload
	if data.Configuration != "" {
		// ini isi field configuration
		dataByte, err := base64.StdEncoding.DecodeString(data.Configuration)
		if err != nil {
			return res, err
		}
		err = yaml.Unmarshal(dataByte, &res)
		if err != nil {
			return res, err
		}
	}

	return res, nil
}

func MergeObject(object ...interface{}) (merge interface{}) {
	tempData := ""

	for k, v := range object {
		if v == nil {
			continue
		}
		dataByte, _ := json.Marshal(v)
		normalized := string(dataByte[1 : len(string(dataByte))-1])
		if normalized != "" {
			if k > 0 {
				tempData += "," + string(normalized)
			} else {
				tempData += string(normalized)
			}
		}
	}
	savedData := "{" + tempData + "}"
	e := json.Unmarshal([]byte(savedData), &merge)
	if e != nil {
		//Append
		appendObject := []interface{}{}
		for _, v := range object {
			appendObject = append(appendObject, v)
		}
		merge = appendObject
	}
	return merge
}

func LogJSON(objs interface{}) string {
	empJSON, err := jsonExt.Marshal(objs)
	if err != nil {
		log.Fatalf("%s", err.Error())
	}
	return string(empJSON)
}

func DownloadFile(URL string) ([]byte, error) {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	req, err := http.NewRequest(http.MethodGet, URL, nil)
	if err != nil {
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	} else if res.StatusCode != 200 {
		return nil, errors.New(res.Status)
	}
	defer res.Body.Close()

	result, err := ioutil.ReadAll(res.Body)
	return result, err
}

// InjectTraceContext injects OpenTelemetry trace context into RabbitMQ message headers
func InjectTraceContext(ctx context.Context, message *cesmodel.MessageQueue) {
	if !config.TelemetryConfig.TelemetryEnabled || telemetry.Tracer == nil {
		return
	}

	// Initialize headers map if it doesn't exist
	if message.Headers == nil {
		message.Headers = make(map[string]interface{})
	}

	// Convert headers to proper map type
	var headersMap map[string]interface{}
	if hm, ok := message.Headers.(map[string]interface{}); ok {
		headersMap = hm
	} else {
		headersMap = make(map[string]interface{})
		message.Headers = headersMap
	}

	// Create a propagation carrier from the headers
	carrier := make(propagation.MapCarrier)
	for key, value := range headersMap {
		if strValue, ok := value.(string); ok {
			carrier[key] = strValue
		}
	}

	// Inject trace context into carrier using global propagator
	otel.GetTextMapPropagator().Inject(ctx, carrier)

	// Copy back to message headers
	for key, value := range carrier {
		headersMap[key] = value
	}
}
