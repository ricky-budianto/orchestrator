package utils

import (
	"bytes"
	"crypto/hmac"
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"text/template"
	"time"
	"unicode"
	"unsafe"

	"github.com/dgrijalva/jwt-go/v4"
	"github.com/google/uuid"
	"github.com/soluixdeveloper/ces-orchestratorService/app/model"
	"github.com/soluixdeveloper/ces-utilities/v2/ceslogger"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
var lettersPassword = []rune("!@#$%^&*-_=+|,.?abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

// WhiteSpaceTrimmer removes white spaces.
func WhiteSpaceTrimmer(str string) string {
	var b strings.Builder
	b.Grow(len(str))
	for _, ch := range str {
		if !unicode.IsSpace(ch) {
			b.WriteRune(ch)
		}
	}
	return b.String()
}

// TimestampNow returns current timestamp.
// If format is not speciefied, it will return timestamp in RFC3339 format.
func TimestampNow(UTC bool, format string) (time.Time, string) {
	var timestampNow time.Time
	if UTC {
		timestampNow = time.Now().UTC()
	} else {
		timestampNow = time.Now().Local()
	}

	if format != "" {
		return timestampNow, timestampNow.Format(format)
	} else {
		return timestampNow, timestampNow.Format(time.RFC3339)
	}
}

// AnyToJsonStr converts any data type to JSON string if the data type itself supported
// to be converted to JSON.
func AnyToJsonStr(data interface{}) string {
	switch data := data.(type) {
	case []byte:
		return string(data)
	default:
		dataByte, err := json.Marshal(data)
		if err != nil {
			return fmt.Sprintf("%v", data)
		}
		return string(dataByte)
	}
}

// AnyToMapStringInterface converts any data type to map[string]interface{}
func AnyToMapStringInterface(data interface{}) map[string]interface{} {
	var result map[string]interface{}
	dataByte, _ := json.Marshal(data)
	json.Unmarshal(dataByte, &result)

	return result
}

// StructIterator iterates struct
func StructIterator(s interface{}) {
	v := reflect.ValueOf(s)
	typeOfS := v.Type()

	for i := 0; i < v.NumField(); i++ {
		fmt.Printf("Field: %s\tValue: %v\n", typeOfS.Field(i).Name, v.Field(i).Interface())
	}
}

// DecodeJWT decodes JWT
func DecodeJWT(accessToken string) (jwt.MapClaims, error) {
	decoded, _ := jwt.Parse(accessToken, nil)

	claims, ok := decoded.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func ParseTemplateToString(tmplt string, data interface{}) (string, error) {

	var err error
	buf := new(bytes.Buffer)
	t, _ := template.New("template").Parse(tmplt)

	if err = t.Execute(buf, data); err != nil {
		return buf.String(), err
	}

	return buf.String(), err
}

func ExpiredTime(day int, hour int, minute int) string {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	t := time.Now()

	// newDate := t.Add(time.Hour * 24 * 1 * time.Duration(1))
	newDate := time.Date(t.Year(), t.Month(), t.Day()+day, t.Hour()+hour, t.Minute()+minute, t.Second(), t.Nanosecond(), loc)

	return newDate.Format(time.RFC3339)
}

func GeneratePassword() string {
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, 12)
	for i := range b {
		b[i] = lettersPassword[rand.Intn(len(letters))]
	}
	return string(b)
}

func ParseTemplate(subject string, templates string, data interface{}) *bytes.Buffer {

	t, _ := template.New("test").Parse(templates)
	var body bytes.Buffer
	headers := "MIME-version: 1.0;\nContent-Type: text/html;"
	body.Write([]byte(fmt.Sprintf("Subject: "+subject+"\n%s\n\n", headers)))

	t.Execute(&body, data)

	return &body
}

func BasicAuth(c string) model.BasicAuth {
	keyAuth := strings.Split(c, " ")
	decodeKeyAuth, _ := base64.StdEncoding.DecodeString(keyAuth[1])
	user := strings.Split(string(decodeKeyAuth), ":")
	var auth model.BasicAuth
	if len(user) > 1 {
		auth.Username = user[0]
		auth.Password = user[1]
	} else {
		return auth
	}
	return auth
}

func ParseFloat(value interface{}) float64 {
	logging := ceslogger.Logger{}

	var v string
	logging.LogDebug("value", value, reflect.TypeOf(value).Kind())
	if reflect.TypeOf(value).Kind() == reflect.String {
		v = value.(string)
	} else if reflect.TypeOf(value).Kind() == reflect.Float64 {
		v = strconv.Itoa(int(value.(float64)))
	} else if reflect.TypeOf(value).Kind() == reflect.Float64 {
		v = strconv.Itoa(value.(int))
	}

	// PARSE FLOAT64
	if s, err := strconv.ParseFloat(v, 64); err == nil {
		fmt.Printf("%.2f\n", s)
		return s
	} else {
		logging.LogDebug("convert string to flaot64", v, err)
	}

	// PARSE FLOAT32
	if s, err := strconv.ParseFloat(v, 32); err == nil {
		fmt.Printf("%.2f\n", s)
		return s
	} else {
		logging.LogDebug("convert string to flaot32", v, err)
	}
	return 0
}

func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

func RandomString(l int) string {
	bytes := make([]byte, l)
	for i := 0; i < l; i++ {
		bytes[i] = byte(randInt(65, 90))
	}
	return string(bytes)
}

func UUIDgenerator() string {
	return uuid.New().String()
}

// B2s converts byte to string
func B2s(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func ExpiredTimeRaw(day int, hour int, minute int) (time.Time, string) {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	t := time.Now()

	// newDate := t.Add(time.Hour * 24 * 1 * time.Duration(1))
	newDate := time.Date(t.Year(), t.Month(), t.Day()+day, t.Hour()+hour, t.Minute()+minute, t.Second(), t.Nanosecond(), loc)

	return newDate, newDate.Format(time.RFC3339)
}

func HMAC_SHA256(secret_key, value string) string {
	h := hmac.New(sha256.New, []byte(secret_key))

	// Write Data to it
	h.Write([]byte(value))

	// Get result and encode as hexadecimal string
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func ReplaceDayName(input, english, indonesian string) string {
	return strings.Replace(input, english, indonesian, -1)
}

func ReplaceMonthName(input, english, indonesian string) string {
	return strings.Replace(input, english, indonesian, -1)
}

func IsSameAsCurrentDate(inputDateString string) (bool, error) {
	layout := MTLDateLayout
	inputTime, err := time.Parse(layout, inputDateString)
	if err != nil {
		return false, err
	}

	currentTime := time.Now()

	return inputTime.Year() == currentTime.Year() &&
		inputTime.Month() == currentTime.Month() &&
		inputTime.Day() == currentTime.Day(), nil
}

func GenRandomOTP(length int) string {
	chars := "1234567890"
	return GenerateRandom(length, chars)
}

func GenerateRandom(length int, obj string) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length)
	for i := range b {
		num, err := crand.Int(crand.Reader, big.NewInt(int64(len(obj))))
		if err != nil {
			return ""
		}
		b[i] = obj[num.Int64()]
	}
	return string(b)
}

func ConvertByKeyName(keyName, input, def string, mappinglist []model.MappingTable) (result string) {

	for _, row := range mappinglist {
		if row.KeyName == keyName {
			if row.InputValue == input && row.OutputValue != "" {
				result = row.OutputValue
				break
			}
		}
	}
	if result == "" {
		result = def
	}

	return
}
