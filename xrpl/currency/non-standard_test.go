package currency

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertStringToHex(t *testing.T) {
	// Test with token length = 0
	t.Run("Empty token", func(t *testing.T) {
		token := ""

		expectedOutput := ""
		output := ConvertStringToHex(token)
		assert.Equal(t, expectedOutput, output)
	})

	// Test with token length = 1
	t.Run("Single character token", func(t *testing.T) {
		token := "A"

		expectedOutput := "A"
		output := ConvertStringToHex(token)
		assert.Equal(t, expectedOutput, output)
	})

	// Test with token length = 2
	t.Run("Two character token", func(t *testing.T) {
		token := "AB"

		expectedOutput := "AB"
		output := ConvertStringToHex(token)
		assert.Equal(t, expectedOutput, output)
	})

	// Test with token length = 3
	t.Run("Three character token", func(t *testing.T) {
		token := "ABC"

		expectedOutput := "ABC"
		output := ConvertStringToHex(token)
		assert.Equal(t, expectedOutput, output)
	})

	// Test with token length > 3
	t.Run("Long token", func(t *testing.T) {
		token := "ABCDE"

		expectedOutput := "4142434445" + "000000000000000000000000000000"
		output := ConvertStringToHex(token)
		assert.Equal(t, expectedOutput, output)
	})
}

func TestConvertHexToString(t *testing.T) {
	// test decoding a hex with the Nonstandard Currency Codes
	t.Run("Hex with Nonstandard Currency Codes", func(t *testing.T) {
		hex := "41424344" + "00000000000000000000000000000000"

		expectedOutput := "ABCD"
		output, _ := ConvertHexToString(hex)
		assert.Equal(t, expectedOutput, output)
	})

	// test decoding a hex without the Nonstandard Currency Codes
	t.Run("Hex without Nonstandard Currency Codes", func(t *testing.T) {
		hex := "41424344"

		expectedOutput := "ABCD"
		output, _ := ConvertHexToString(hex)
		assert.Equal(t, expectedOutput, output)
	})

	// test decoding an empty hex
	t.Run("Empty Hex", func(t *testing.T) {
		hex := ""

		expectedOutput := ""
		output, _ := ConvertHexToString(hex)
		assert.Equal(t, expectedOutput, output)
	})

	// test with invalid hex
	t.Run("Invalid hex", func(t *testing.T) {
		hex := "41424344G"

		_, err := ConvertHexToString(hex)
		require.Error(t, err)
	})
}
