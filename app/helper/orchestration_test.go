package helper

import (
	"reflect"
	"testing"

	"github.com/magiconair/properties/assert"

	ceslogger "github.com/soluixdeveloper/ces-utilities/v2/ceslogger"
)

// go test -v -timeout 30s -run ^TestIbridgeHelpers$ github.com/soluixdeveloper/ces-orchestratorService/app/helper
func TestIbridgeHelpers(t *testing.T) {
	t.Log("~TestIbridgeHelpers")
	ceslogger.InitLogger(
		"-1",
		"LOCAL",
		"Test",
		"",
		"1.0.0")
	logging := ceslogger.Logger{}

	actual1 := ibridgeHelpers("${{ibridge.Helpers.StringConcat}}(abc,123,def)", map[string]interface{}{}, logging)
	t.Log("actual1:", actual1)
	assert.Equal(t, "abc123def", actual1)

	actual2 := ibridgeHelpers("${{ibridge.Helpers.ConvertGenderTm}}(LAKI-LAKI)", map[string]interface{}{}, logging)
	t.Log("actual2:", actual2)
	assert.Equal(t, "CUSTOMER_GENDER_MALE", actual2)

	actual3 := ibridgeHelpers("${{ibridge.Helpers.ConvertGenderTm}}(PEREMPUAN)", map[string]interface{}{}, logging)
	t.Log("actual3:", actual3)
	assert.Equal(t, "CUSTOMER_GENDER_FEMALE", actual3)

	actual4 := ibridgeHelpers("${{ibridge.Helpers.ConvertGenderTm}}(LGBT)", map[string]interface{}{}, logging)
	t.Log("actual4:", actual4)
	assert.Equal(t, "CUSTOMER_GENDER_UNKNOWN", actual4)

	actual5 := ibridgeHelpers("${{ibridge.Helpers.ConvertTimestampFormat}}(2025-03-23 20:00:00,2006-01-02 15:04:05,2006-01-02,local)", map[string]interface{}{}, logging)
	t.Log("actual5:", actual5)
	assert.Equal(t, "2025-03-24", actual5)

	actual6 := ibridgeHelpers("${{ibridge.Helpers.ConvertTimestampFormat}}(2025-03-24 03:00:00,2006-01-02 15:04:05,2006-01-02,local)", map[string]interface{}{}, logging)
	t.Log("actual6:", actual6)
	assert.Equal(t, "2025-03-24", actual6)

	actual7 := ibridgeHelpers("${{ibridge.Helpers.GetLength}}([],0)", map[string]interface{}{}, logging)
	t.Log("actual7:", actual7)
	assert.Equal(t, 0, actual7)

	// Test AnyToString
	actual8 := ibridgeHelpers("${{ibridge.Helpers.AnyToString}}(test123)", map[string]interface{}{}, logging)
	t.Log("actual8:", actual8)
	assert.Equal(t, "test123", actual8)

	// Test StringTrimLength - from left
	actual9 := ibridgeHelpers("${{ibridge.Helpers.StringTrimLength}}(abcdefgh,5,left)", map[string]interface{}{}, logging)
	t.Log("actual9:", actual9)
	assert.Equal(t, "abcde", actual9)

	// Test StringTrimLength - from right
	actual10 := ibridgeHelpers("${{ibridge.Helpers.StringTrimLength}}(abcdefgh,5,right)", map[string]interface{}{}, logging)
	t.Log("actual10:", actual10)
	assert.Equal(t, "defgh", actual10)

	// Test StringPaddingZero - left
	actual11 := ibridgeHelpers("${{ibridge.Helpers.StringPaddingZero}}(123,6,left)", map[string]interface{}{}, logging)
	t.Log("actual11:", actual11)
	assert.Equal(t, "000123", actual11)

	// Test StringPaddingZero - right
	actual12 := ibridgeHelpers("${{ibridge.Helpers.StringPaddingZero}}(123,6,right)", map[string]interface{}{}, logging)
	t.Log("actual12:", actual12)
	assert.Equal(t, "123000", actual12)

	// Test StringPaddingSpace - left
	actual13 := ibridgeHelpers("${{ibridge.Helpers.StringPaddingSpace}}(test,8,left)", map[string]interface{}{}, logging)
	t.Log("actual13:", actual13)
	assert.Equal(t, "    test", actual13)

	// Test StringPaddingSpace - right
	actual14 := ibridgeHelpers("${{ibridge.Helpers.StringPaddingSpace}}(test,8,right)", map[string]interface{}{}, logging)
	t.Log("actual14:", actual14)
	assert.Equal(t, "test    ", actual14)

	// Test SetDefaultString - with value
	actual15 := ibridgeHelpers("${{ibridge.Helpers.SetDefaultString}}(myvalue,defaultval)", map[string]interface{}{}, logging)
	t.Log("actual15:", actual15)
	assert.Equal(t, "myvalue", actual15)

	// Test SetDefaultString - without value
	actual16 := ibridgeHelpers("${{ibridge.Helpers.SetDefaultString}}(,defaultval)", map[string]interface{}{}, logging)
	t.Log("actual16:", actual16)
	assert.Equal(t, "defaultval", actual16)

	// Test Sum
	actual17 := ibridgeHelpers("${{ibridge.Helpers.Sum}}(10,20,30.5)", map[string]interface{}{}, logging)
	t.Log("actual17:", actual17)
	assert.Equal(t, 60.5, actual17)

	// Test TrimSpace
	actual18 := ibridgeHelpers("${{ibridge.Helpers.TrimSpace}}(  hello world  )", map[string]interface{}{}, logging)
	t.Log("actual18:", actual18)
	assert.Equal(t, "hello world", actual18)

	// Test TrimPrefix
	actual19 := ibridgeHelpers("${{ibridge.Helpers.TrimPrefix}}(prefixtest,prefix)", map[string]interface{}{}, logging)
	t.Log("actual19:", actual19)
	assert.Equal(t, "test", actual19)

	// Test RemoveStringElement - left
	actual20 := ibridgeHelpers("${{ibridge.Helpers.RemoveStringElement}}(abcdefgh,3,left)", map[string]interface{}{}, logging)
	t.Log("actual20:", actual20)
	assert.Equal(t, "defgh", actual20)

	// Test RemoveStringElement - right
	actual21 := ibridgeHelpers("${{ibridge.Helpers.RemoveStringElement}}(abcdefgh,3,right)", map[string]interface{}{}, logging)
	t.Log("actual21:", actual21)
	assert.Equal(t, "abcde", actual21)

	// Test StringTrimLeft
	actual22 := ibridgeHelpers("${{ibridge.Helpers.StringTrimLeft}}(000123,0)", map[string]interface{}{}, logging)
	t.Log("actual22:", actual22)
	assert.Equal(t, "123", actual22)

	// Test StringTrimRight
	actual23 := ibridgeHelpers("${{ibridge.Helpers.StringTrimRight}}(123000,0)", map[string]interface{}{}, logging)
	t.Log("actual23:", actual23)
	assert.Equal(t, "123", actual23)

	// Test StringReplaceAll
	actual24 := ibridgeHelpers("${{ibridge.Helpers.StringReplaceAll}}(hello.world,.,_)", map[string]interface{}{}, logging)
	t.Log("actual24:", actual24)
	assert.Equal(t, "hello_world", actual24)

	// Test AddLuhnAlgorithm
	actual25 := ibridgeHelpers("${{ibridge.Helpers.AddLuhnAlgorithm}}(123456)", map[string]interface{}{}, logging)
	t.Log("actual25:", actual25)
	assert.Equal(t, "1234566", actual25)

	// Test Multiply
	actual26 := ibridgeHelpers("${{ibridge.Helpers.Multiply}}(10,2.5)", map[string]interface{}{}, logging)
	t.Log("actual26:", actual26)
	assert.Equal(t, 25.0, actual26)

	// Test MathIntDiv
	actual27 := ibridgeHelpers("${{ibridge.Helpers.MathIntDiv}}(13,3)", map[string]interface{}{}, logging)
	t.Log("actual27:", actual27)
	assert.Equal(t, 4, actual27)

	// Test MathIntMod
	actual28 := ibridgeHelpers("${{ibridge.Helpers.MathIntMod}}(13,3)", map[string]interface{}{}, logging)
	t.Log("actual28:", actual28)
	assert.Equal(t, 1, actual28)

	// Test FloatRounding
	actual29 := ibridgeHelpers("${{ibridge.Helpers.FloatRounding}}(3.14159,2)", map[string]interface{}{}, logging)
	t.Log("actual29:", actual29)
	assert.Equal(t, 3.14, actual29)

	// Test MathDivision
	actual30 := ibridgeHelpers("${{ibridge.Helpers.MathDivision}}(10,4,2)", map[string]interface{}{}, logging)
	t.Log("actual30:", actual30)
	assert.Equal(t, 2.5, actual30)

	// Test SubtractDateTime - days
	actual31 := ibridgeHelpers("${{ibridge.Helpers.SubtractDateTime}}(days,2006-01-02,2025-01-10,2025-01-05)", map[string]interface{}{}, logging)
	t.Log("actual31:", actual31)
	assert.Equal(t, "5", actual31)

	// Test ToLower
	actual32 := ibridgeHelpers("${{ibridge.Helpers.ToLower}}(HELLO WORLD)", map[string]interface{}{}, logging)
	t.Log("actual32:", actual32)
	assert.Equal(t, "hello world", actual32)

	// Test ToTitle
	actual33 := ibridgeHelpers("${{ibridge.Helpers.ToTitle}}(hello world)", map[string]interface{}{}, logging)
	t.Log("actual33:", actual33)
	assert.Equal(t, "Hello World", actual33)

	// Test CensorAccountNumber
	actual34 := ibridgeHelpers("${{ibridge.Helpers.CensorAccountNumber}}(8888071035)", map[string]interface{}{}, logging)
	t.Log("actual34:", actual34)
	assert.Equal(t, "8888****35", actual34)

	// Test CurrencyFormatID
	actual35 := ibridgeHelpers("${{ibridge.Helpers.CurrencyFormatID}}(1234567.89)", map[string]interface{}{}, logging)
	t.Log("actual35:", actual35)
	assert.Equal(t, "1.234.567,89", actual35)

	// Test CurrencyFormatID - simple
	actual36 := ibridgeHelpers("${{ibridge.Helpers.CurrencyFormatID}}(1000)", map[string]interface{}{}, logging)
	t.Log("actual36:", actual36)
	assert.Equal(t, "1.000,00", actual36)

	// Test ComputeCRC32
	actual37 := ibridgeHelpers("${{ibridge.Helpers.ComputeCRC32}}(test)", map[string]interface{}{}, logging)
	t.Log("actual37:", actual37)
	assert.Equal(t, uint32(3632233996), actual37)

	// Test AppendArray - new array
	actual38 := ibridgeHelpers("${{ibridge.Helpers.AppendArray}}(new,item1,item2,item3)", map[string]interface{}{}, logging)
	t.Log("actual38:", actual38)
	result38, ok38 := actual38.([]interface{})
	assert.Equal(t, true, ok38)
	assert.Equal(t, 3, len(result38))
	assert.Equal(t, "item1", result38[0])

	// Test SumArray Float
	actual39 := ibridgeHelpers("${{ibridge.Helpers.SumArray}}(${{test}})", map[string]interface{}{"test": []float64{2, 3, 1.2}}, logging)
	t.Log("actual39:", actual39)
	assert.Equal(t, 6.2, actual39)

	// Test SumArray Str
	actual40 := ibridgeHelpers("${{ibridge.Helpers.SumArray}}(${{test}})", map[string]interface{}{"test": []string{"2", "3", "1.2"}}, logging)
	t.Log("actual40:", actual40)
	assert.Equal(t, 6.2, actual40)
}

// go test -v -timeout 30s -run ^TestNotEqualOperator$ github.com/soluixdeveloper/ces-orchestratorService/app/helper
func TestNotEqualOperator(t *testing.T) {
	t.Log("~TestNotEqualOperator")
	ceslogger.InitLogger(
		"-1",
		"LOCAL",
		"Test",
		"",
		"1.0.0")

	// Test NotEqual with different strings - should be NOT equal (valid = true)
	testCase1 := testNotEqual("hello", "world")
	t.Log("testCase1 (hello != world):", testCase1)
	assert.Equal(t, true, testCase1)

	// Test NotEqual with same strings - should be equal (valid = false)
	testCase2 := testNotEqual("hello", "hello")
	t.Log("testCase2 (hello != hello):", testCase2)
	assert.Equal(t, false, testCase2)

	// Test NotEqual with different numbers - should be NOT equal (valid = true)
	testCase3 := testNotEqual(123, 456)
	t.Log("testCase3 (123 != 456):", testCase3)
	assert.Equal(t, true, testCase3)

	// Test NotEqual with same numbers - should be equal (valid = false)
	testCase4 := testNotEqual(123, 123)
	t.Log("testCase4 (123 != 123):", testCase4)
	assert.Equal(t, false, testCase4)

	// Test NotEqual with different types (int vs float64) - should be NOT equal (valid = true)
	testCase5 := testNotEqual(123, 123.0)
	t.Log("testCase5 (123 != 123.0):", testCase5)
	assert.Equal(t, true, testCase5)

	// Test NotEqual with nil vs non-nil - should be NOT equal (valid = true)
	testCase6 := testNotEqual(nil, "something")
	t.Log("testCase6 (nil != something):", testCase6)
	assert.Equal(t, true, testCase6)

	// Test NotEqual with both nil - should be equal (valid = false)
	testCase7 := testNotEqual(nil, nil)
	t.Log("testCase7 (nil != nil):", testCase7)
	assert.Equal(t, false, testCase7)

	// Test NotEqual with maps with different content - should be NOT equal (valid = true)
	map1 := map[string]interface{}{"key": "value1"}
	map2 := map[string]interface{}{"key": "value2"}
	testCase8 := testNotEqual(map1, map2)
	t.Log("testCase8 (map1 != map2):", testCase8)
	assert.Equal(t, true, testCase8)

	// Test NotEqual with maps with same content - should be equal (valid = false)
	map3 := map[string]interface{}{"key": "value"}
	map4 := map[string]interface{}{"key": "value"}
	testCase9 := testNotEqual(map3, map4)
	t.Log("testCase9 (map3 != map4 with same content):", testCase9)
	assert.Equal(t, false, testCase9)

	// Test NotEqual with arrays with different content - should be NOT equal (valid = true)
	arr1 := []interface{}{"a", "b", "c"}
	arr2 := []interface{}{"x", "y", "z"}
	testCase10 := testNotEqual(arr1, arr2)
	t.Log("testCase10 (arr1 != arr2):", testCase10)
	assert.Equal(t, true, testCase10)

	// Test NotEqual with arrays with same content - should be equal (valid = false)
	arr3 := []interface{}{"a", "b", "c"}
	arr4 := []interface{}{"a", "b", "c"}
	testCase11 := testNotEqual(arr3, arr4)
	t.Log("testCase11 (arr3 != arr4 with same content):", testCase11)
	assert.Equal(t, false, testCase11)

	// Test NotEqual with both nil - should be equal (valid = false)
	testCase12 := testNotEqual("0", 0)
	t.Log("testCase12 (\"0\" != 0):", testCase12)
	assert.Equal(t, true, testCase12)
}

// testNotEqual simulates the NotEqual operator logic from orchestration.go:855-858
// Returns true if values are NOT equal (valid condition), false if they are equal
func testNotEqual(variableCheck, comparator interface{}) bool {
	valid := true
	// This is the exact logic from orchestration.go line 855-858
	if reflect.DeepEqual(variableCheck, comparator) {
		valid = false
	}
	return valid
}
