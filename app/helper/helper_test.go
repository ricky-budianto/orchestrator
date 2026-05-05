package helper

import (
	"fmt"
	"github.com/magiconair/properties/assert"
	"testing"
)

func TestCalculateCheckDigit(t *testing.T) {
	testCases := []struct {
		testCase string
		expected int
	}{
		{testCase: "7992739871", expected: 3},
		{testCase: "144124151412444", expected: 0},
		{testCase: "4523524362334534532452", expected: 3},
		{testCase: "12343216242", expected: 8},
		{testCase: "094353849789", expected: 3},
		{testCase: "5704", expected: 2},
		{testCase: "341", expected: 8},
		{testCase: "1", expected: 8},
		{testCase: "10", expected: 9},
		{testCase: "01", expected: 8},
		{testCase: "0101010101010", expected: 4},
		{testCase: "11111111111111", expected: 9},
		{testCase: "000000000000", expected: 0},
		{testCase: "9999999999999", expected: 3},
		{testCase: "09", expected: 1},
		{testCase: "90", expected: 1},
	}
	for i, test := range testCases {
		testName := fmt.Sprintf("#%d", i+1)
		t.Run(testName, func(t *testing.T) {
			assert.Equal(t, test.expected, CalculateCheckDigit(test.testCase))
		})
	}
}
