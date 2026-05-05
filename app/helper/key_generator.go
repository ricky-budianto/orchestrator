package helper

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
var lettersOTP = []rune("1234567890")
var lettersPassword = []rune("!@#$%^&*()-_=+[]|<>?abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

var (
	CharsNumber = []rune("1234567890")
	CharsLower  = []rune("abcdefghijklmnopqrstuvwxyz")
	CharsUpper  = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	charsMap    = map[string][]rune{
		"true:false:false": CharsNumber,
		"false:true:false": CharsLower,
		"false:false:true": CharsUpper,
		"true:true:false":  append(append([]rune{}, CharsNumber...), CharsLower...),
		"true:false:true":  append(append([]rune{}, CharsNumber...), CharsUpper...),
		"false:true:true":  append(append([]rune{}, CharsLower...), CharsUpper...),
		"true:true:true":   append(append(append([]rune{}, CharsNumber...), CharsLower...), CharsUpper...),
	}
)

func GeneratePassword() string {
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, 12)
	for i := range b {
		b[i] = lettersPassword[rand.Intn(len(letters))]
	}
	return string(b)
}
func GenerateOTP() string {
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, 6)
	for i := range b {
		b[i] = lettersOTP[rand.Intn(len(lettersOTP))]
	}
	return string(b)
}
func GenerateNumber(length int) string {
	if length < 1 {
		length = 12
	}
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, length)
	for i := range b {
		b[i] = lettersOTP[rand.Intn(len(lettersOTP))]
	}
	return string(b)
}
func GenerateHexadecimal(length int) string {
	if length < 1 {
		length = 12
	}
	// Calculate the number of bytes needed to generate the SHA
	numBytes := length / 2

	// Generate random bytes
	randomBytes := make([]byte, numBytes)
	rand.Read(randomBytes)

	// Compute the SHA hash
	shaHash := sha1.Sum(randomBytes)

	// Convert the SHA hash to a hex string
	shaString := hex.EncodeToString(shaHash[:])

	// Trim the string to the desired length
	if len(shaString) > length {
		shaString = shaString[:length]
	}
	return shaString
}
func GenerateString(length int) string {
	if length < 1 {
		length = 12
	}
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, length)
	for i := range b {
		b[i] = lettersPassword[rand.Intn(len(letters))]
	}
	return string(b)
}
func GenerateKey() string {
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, 12)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
func GenerateID() string {
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, 9)
	for i := range b {
		b[i] = lettersOTP[rand.Intn(len(lettersOTP))]
	}
	return string(b)
}

func GenerateReferenceId(length int, isNumeric, isLowercase, isUppercase bool) string {
	if length <= 0 {
		return ""
	}

	chars, ok := charsMap[fmt.Sprintf("%t:%t:%t", isNumeric, isLowercase, isUppercase)]
	if !ok {
		return ""
	}
	charsLen := len(chars)

	sb := strings.Builder{}
	for i := 0; i < length; i++ {
		_, err := sb.WriteRune(
			chars[rand.Intn(charsLen)],
		)
		if err != nil {
			return ""
		}
	}
	return sb.String()
}

func Convert(i interface{}) interface{} {
	switch x := i.(type) {
	case map[interface{}]interface{}:
		m2 := map[string]interface{}{}
		for k, v := range x {
			m2[k.(string)] = Convert(v)
		}
		return m2
	case []interface{}:
		for i, v := range x {
			x[i] = Convert(v)
		}
	}
	return i
}
func UUIDgenerator() string {
	return uuid.New().String()
}
