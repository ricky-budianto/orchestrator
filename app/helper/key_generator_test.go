package helper

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func stringContainsOnly(input string, whitelistedChars ...[]rune) bool {
	whitelistedCharSet := map[rune]struct{}{}
	for _, chars := range whitelistedChars {
		for _, whiteChr := range chars {
			whitelistedCharSet[whiteChr] = struct{}{}
		}
	}

	for _, chr := range input {
		if _, ok := whitelistedCharSet[chr]; !ok {
			// input mengandung karakter yang tidak di-whitelist
			return false
		}
	}
	return true
}

func TestGenerateReferenceId(t *testing.T) {
	t.Run("test length", func(t *testing.T) {
		tests := []int{
			0, 1, 50, 100,
		}

		for _, idLength := range tests {
			t.Run(
				fmt.Sprintf("generate %d chars", idLength),
				func(t *testing.T) {
					// ketika
					result := GenerateReferenceId(idLength, true, false, false)

					// maka
					require.Len(t, result, idLength)
				},
			)
		}
	})

	t.Run("empty string when given negative length", func(t *testing.T) {
		// ketika
		result := GenerateReferenceId(-1, true, false, false)

		// maka
		require.Len(t, result, 0)
	})

	t.Run("empty string when no valid character type is given", func(t *testing.T) {
		// ketika
		result := GenerateReferenceId(100, false, false, false)

		// maka
		require.Len(t, result, 0)
	})

	t.Run("single character type", func(t *testing.T) {
		t.Run("number characters only", func(t *testing.T) {
			// ketika
			result := GenerateReferenceId(200, true, false, false)

			// maka
			require.True(t, stringContainsOnly(result, CharsNumber), "should contain number chars only")
		})
		t.Run("lowercase characters only", func(t *testing.T) {
			// ketika
			result := GenerateReferenceId(200, false, true, false)

			// maka
			require.True(t, stringContainsOnly(result, CharsLower), "should contain lowercase chars only")
		})
		t.Run("uppercase characters only", func(t *testing.T) {
			// ketika
			result := GenerateReferenceId(200, false, false, true)

			// maka
			require.True(t, stringContainsOnly(result, CharsUpper), "should contain uppercase chars only")
		})
	})

	t.Run("multiple character types", func(t *testing.T) {
		t.Run("number+lowercase characters", func(t *testing.T) {
			// ketika
			result := GenerateReferenceId(200, true, true, false)

			// maka
			require.True(t, stringContainsOnly(result, CharsNumber, CharsLower), "should only contain number & lower chars only")
		})
		t.Run("number+uppercase characters", func(t *testing.T) {
			// ketika
			result := GenerateReferenceId(200, true, false, true)

			// maka
			require.True(t, stringContainsOnly(result, CharsNumber, CharsUpper), "should only contain number & upper chars only")
		})
		t.Run("lowercase+uppercase characters", func(t *testing.T) {
			// ketika
			result := GenerateReferenceId(200, false, true, true)

			// maka
			require.True(t, stringContainsOnly(result, CharsLower, CharsUpper), "should only contain lower & upper chars only")
		})
		t.Run("all characters", func(t *testing.T) {
			// ketika
			result := GenerateReferenceId(200, true, true, true)

			// maka
			require.True(t, stringContainsOnly(result, CharsNumber, CharsLower, CharsUpper), "should only contain number, lower, & upper chars only")
		})
	})
}
