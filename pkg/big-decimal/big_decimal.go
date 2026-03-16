// Package bigdecimal provides arbitrary-precision decimal arithmetic operations for financial calculations.
package bigdecimal

import (
	"math/big"
	"regexp"
	"strconv"
	"strings"
)

const (
	// AllowedCharacters defines the set of characters permitted in a BigDecimal string.
	AllowedCharacters = "0123456789.-eE"
	// BigDecRegEx is the regular expression that a valid BigDecimal string must match.
	BigDecRegEx = "-?(?:[0|1-9]\\d*)(?:\\.\\d+)?(?:[eE][+\\-]?\\d+)?"
	// Precision specifies the bit precision used for internal big.Float calculations.
	Precision = 512
)

// BigDecimal represents a high-precision decimal value with scale, precision, and sign.
type BigDecimal struct {
	Scale         int
	Precision     int
	UnscaledValue string
	Sign          int // 1 for negative, 0 for positive
}

// GetScaledValue returns the decimal as a string without scientific notation, scaled by its Scale.
func (bd *BigDecimal) GetScaledValue() string {
	if bd.UnscaledValue == "" {
		return "0"
	}

	// Use SetPrec to maintain full precision
	unscaled := new(big.Float).SetPrec(Precision) // Use high precision to avoid scientific notation
	unscaled, _ = unscaled.SetString(bd.UnscaledValue)

	scalingFactor := new(big.Float).SetPrec(Precision).SetFloat64(1)
	for i := 0; i < abs(bd.Scale); i++ {
		scalingFactor.Mul(scalingFactor, big.NewFloat(10))
	}

	var scaledValue *big.Float
	if bd.Scale >= 0 {
		scaledValue = new(big.Float).SetPrec(Precision).Mul(unscaled, scalingFactor)
	} else {
		scaledValue = new(big.Float).SetPrec(Precision).Quo(unscaled, scalingFactor)
	}

	if bd.Sign == 1 {
		scaledValue.Neg(scaledValue)
	}

	// Force format without scientific notation
	return strings.TrimSuffix(strings.TrimRight(scaledValue.Text('f', abs(bd.Scale)), "0"), ".")
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// NewBigDecimal creates a BigDecimal from the given string, validating format and scale.
func NewBigDecimal(value string) (bd *BigDecimal, err error) {
	// check if the value string contains only allowed characters
	if !bigDecimalRegEx(value) {
		return nil, ErrInvalidCharacter{
			Allowed: AllowedCharacters,
		}
	}

	v := strings.ToLower(value)
	bd = new(BigDecimal)

	// check if the value is negative and set the sign accordingly
	bd.Sign, v = handleSign(v)

	// check if the value contains the 'e' character and split the string into prefix and suffix accordingly
	p, s, eFound := strings.Cut(v, "e")

	// if the prefix without trailing & leading zeros is empty or only contains a decimal character, return an error
	trimP := strings.Trim(p, "0")
	if trimP == "" || trimP == "." {
		return nil, ErrInvalidZeroValue
	}

	// if the value contains the 'e' character, call the appropriate function to get the scale and unscaled value
	if eFound {
		bd.Scale, bd.UnscaledValue = getScaleAndUnscaledValWithE(p, s)
	} else {
		bd.Scale, bd.UnscaledValue = getScaleAndUnscaledValNoE(p, s)
	}

	if bd.UnscaledValue == "" {
		return nil, ErrInvalidZeroValue
	}

	bd.Precision = len(bd.UnscaledValue)
	return
}

func getScaleAndUnscaledValNoE(p, _ string) (sc int, uv string) {
	// check if the value contains a decimal character and split the string into prefix and suffix accordingly
	decP, decS, decFound := strings.Cut(p, ".")
	if decFound {
		return valHasDecimal(0, decP, decS)
	}

	return valNoDecimalNoE(0, p, decP)
}

func getScaleAndUnscaledValWithE(p, s string) (sc int, uv string) {
	// check if the value contains a decimal character and split the string into prefix and suffix accordingly
	decP, decS, decFound := strings.Cut(p, ".")
	sc, err := strconv.Atoi(s)
	if err != nil {
		return 0, ""
	}
	if decFound {
		return valHasDecimal(sc, decP, decS)
	}
	return valNoDecimalHasE(sc, p, decP)
}

func valHasDecimal(scale int, decP, decS string) (sc int, uv string) {
	uv = strings.Trim((decP + decS), "0")
	sc = scale - len(strings.TrimRight(decS, "0"))
	if strings.TrimRight(decS, "0") == "" {
		sc = scale + len(strings.TrimLeft(decP, "0")) - len(uv)
	}
	return
}

func valNoDecimalNoE(_ int, prefix, decP string) (sc int, uv string) {
	uv = strings.Trim(decP, "0")
	sc = len(prefix) - len(strings.TrimRight(decP, "0"))
	return
}

func valNoDecimalHasE(scale int, prefix, _ string) (sc int, uv string) {
	uv = strings.Trim(prefix, "0")
	sc = scale + len(strings.TrimLeft(prefix, "0")) - len(uv)
	return
}

func handleSign(value string) (int, string) {
	if after, ok := strings.CutPrefix(value, "-"); ok {
		return 1, after
	}
	return 0, value
}

func bigDecimalRegEx(value string) bool {
	r := regexp.MustCompile(BigDecRegEx)
	m := r.FindAllString(value, -1)
	return len(m) == 1
}
