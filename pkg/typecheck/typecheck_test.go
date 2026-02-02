package typecheck

import "testing"

func TestIsString(t *testing.T) {
	tests := []struct {
		name string
		str  interface{}
		want bool
	}{
		{
			name: "pass - Valid string",
			str:  "Hello, World!",
			want: true,
		},
		{
			name: "pass - Invalid string",
			str:  42,
			want: false,
		},
		{
			name: "pass - Empty string",
			str:  "",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsString(tt.str); got != tt.want {
				t.Errorf("IsString(%v) = %v, want %v", tt.str, got, tt.want)
			}
		})
	}
}
func TestIsUint32(t *testing.T) {
	tests := []struct {
		name string
		num  interface{}
		want bool
	}{
		{
			name: "pass - Valid uint32",
			num:  uint32(42),
			want: true,
		},
		{
			name: "pass - Invalid uint32",
			num:  42,
			want: false,
		},
		{
			name: "pass - Valid uint64",
			num:  uint64(42),
			want: false,
		},
		{
			name: "pass - Valid int",
			num:  int(42),
			want: false,
		},
		{
			name: "pass - Valid uint",
			num:  uint(42),
			want: false,
		},
		{
			name: "pass - Valid bool",
			num:  true,
			want: false,
		},
		{
			name: "pass - Valid map",
			num:  map[string]interface{}{},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsUint32(tt.num); got != tt.want {
				t.Errorf("IsUint32(%v) = %v, want %v", tt.num, got, tt.want)
			}
		})
	}
}
func TestIsUint64(t *testing.T) {
	tests := []struct {
		name string
		num  interface{}
		want bool
	}{
		{
			name: "pass - Valid uint64",
			num:  uint64(42),
			want: true,
		},
		{
			name: "pass - Invalid uint64",
			num:  42,
			want: false,
		},
		{
			name: "pass - Valid uint32",
			num:  uint32(42),
			want: false,
		},
		{
			name: "pass - Valid int",
			num:  int(42),
			want: false,
		},
		{
			name: "pass - Valid uint",
			num:  uint(42),
			want: false,
		},
		{
			name: "pass - Valid bool",
			num:  true,
			want: false,
		},
		{
			name: "pass - Valid map",
			num:  map[string]interface{}{},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsUint64(tt.num); got != tt.want {
				t.Errorf("IsUint64(%v) = %v, want %v", tt.num, got, tt.want)
			}
		})
	}
}
func TestIsUint(t *testing.T) {
	tests := []struct {
		name string
		num  interface{}
		want bool
	}{
		{
			name: "pass - Valid uint",
			num:  uint(42),
			want: true,
		},
		{
			name: "pass - Invalid uint",
			num:  42,
			want: false,
		},
		{
			name: "pass - Valid uint32",
			num:  uint32(42),
			want: false,
		},
		{
			name: "pass - Valid uint64",
			num:  uint64(42),
			want: false,
		},
		{
			name: "pass - Valid int",
			num:  int(42),
			want: false,
		},
		{
			name: "pass - Valid bool",
			num:  true,
			want: false,
		},
		{
			name: "pass - Valid map",
			num:  map[string]interface{}{},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsUint(tt.num); got != tt.want {
				t.Errorf("IsUint(%v) = %v, want %v", tt.num, got, tt.want)
			}
		})
	}
}
func TestIsBool(t *testing.T) {
	tests := []struct {
		name string
		b    interface{}
		want bool
	}{
		{
			name: "pass - Valid bool",
			b:    true,
			want: true,
		},
		{
			name: "pass - Invalid bool",
			b:    42,
			want: false,
		},
		{
			name: "pass - Invalid bool",
			b:    "true",
			want: false,
		},
		{
			name: "pass - Invalid bool",
			b:    nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsBool(tt.b); got != tt.want {
				t.Errorf("IsBool(%v) = %v, want %v", tt.b, got, tt.want)
			}
		})
	}
}

func TestIsHex(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want bool
	}{
		{
			name: "pass - Valid hexadecimal string",
			s:    "0123456789abcdefABCDEF",
			want: true,
		},
		{
			name: "pass - Invalid hexadecimal string with non-hex characters",
			s:    "0123456789abcdefABCDEFG",
			want: false,
		},
		{
			name: "pass - Invalid hexadecimal string with spaces",
			s:    "0123456789 abcdefABCDEF",
			want: false,
		},
		{
			name: "pass - Invalid hexadecimal string with special characters",
			s:    "0123456789!abcdefABCDEF",
			want: false,
		},
		{
			name: "pass - Empty string",
			s:    "",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsHex(tt.s); got != tt.want {
				t.Errorf("IsValidHex(%q) = %v, want %v", tt.s, got, tt.want)
			}
		})
	}
}
func TestIsInt(t *testing.T) {
	tests := []struct {
		name string
		num  interface{}
		want bool
	}{
		{
			name: "pass - Valid int",
			num:  42,
			want: true,
		},
		{
			name: "pass - Invalid int",
			num:  3.14,
			want: false,
		},
		{
			name: "pass - Invalid int",
			num:  "42",
			want: false,
		},
		{
			name: "pass - Invalid int",
			num:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsInt(tt.num); got != tt.want {
				t.Errorf("IsInt(%v) = %v, want %v", tt.num, got, tt.want)
			}
		})
	}
}

func TestIsFloat64(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want bool
	}{
		{
			name: "pass - Valid float64",
			s:    "3.141592653589793",
			want: true,
		},
		{
			name: "pass - Valid float64 (integer)",
			s:    "42",
			want: true,
		},
		{
			name: "pass - Valid negative float64",
			s:    "-42.0",
			want: true,
		},
		{
			name: "pass - Valid float64 with leading zero",
			s:    "0.123456789",
			want: true,
		},
		{
			name: "pass - Valid negative float64",
			s:    "-3.141592653589793",
			want: true,
		},
		{
			name: "pass - Invalid float64 with multiple decimal points",
			s:    "3.14.15",
			want: false,
		},
		{
			name: "pass - Invalid float64 with non-numeric characters",
			s:    "3.14abc",
			want: false,
		},
		{
			name: "pass - Valid float64 with leading plus sign",
			s:    "+3.141592653589793",
			want: true,
		},
		{
			name: "pass - Invalid float64 with leading minus sign",
			s:    "-",
			want: false,
		},
		{
			name: "pass - Invalid float64 with empty string",
			s:    "",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsFloat64(tt.s); got != tt.want {
				t.Errorf("IsFloat64(%q) = %v, want %v", tt.s, got, tt.want)
			}
		})
	}
}
func TestIsFloat32(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want bool
	}{
		{
			name: "pass - Valid float32",
			s:    "3.14159",
			want: true,
		},
		{
			name: "pass - Valid float32 (integer)",
			s:    "42",
			want: true,
		},
		{
			name: "pass - Valid negative float32",
			s:    "-42.0",
			want: true,
		},
		{
			name: "pass - Valid float32 with leading zero",
			s:    "0.123456",
			want: true,
		},
		{
			name: "pass - Valid negative float32",
			s:    "-3.14159",
			want: true,
		},
		{
			name: "pass - Invalid float32 with multiple decimal points",
			s:    "3.14.15",
			want: false,
		},
		{
			name: "pass - Invalid float32 with non-numeric characters",
			s:    "3.14abc",
			want: false,
		},
		{
			name: "pass - Valid float32 with leading plus sign",
			s:    "+3.14159",
			want: true,
		},
		{
			name: "pass - Invalid float32 with leading minus sign",
			s:    "-",
			want: false,
		},
		{
			name: "pass - Invalid float32 with empty string",
			s:    "",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsFloat32(tt.s); got != tt.want {
				t.Errorf("IsFloat32(%q) = %v, want %v", tt.s, got, tt.want)
			}
		})
	}
}
func TestIsStringNumericUint(t *testing.T) {
	tests := []struct {
		name    string
		s       string
		base    int
		bitSize int
		want    bool
	}{
		{
			name:    "pass - Valid uint string",
			s:       "42",
			base:    10,
			bitSize: 64,
			want:    true,
		},
		{
			name:    "pass - Valid large uint string",
			s:       "18446744073709551615", // Max uint64 value
			base:    10,
			bitSize: 64,
			want:    true,
		},
		{
			name:    "pass - Invalid uint string with negative sign",
			s:       "-42",
			base:    10,
			bitSize: 64,
			want:    false,
		},
		{
			name:    "pass - Invalid uint string with decimal point",
			s:       "42.0",
			base:    10,
			bitSize: 64,
			want:    false,
		},
		{
			name:    "pass - Invalid uint string with non-numeric characters",
			s:       "42abc",
			base:    10,
			bitSize: 64,
			want:    false,
		},
		{
			name:    "pass - Invalid uint string with special characters",
			s:       "42!",
			base:    10,
			bitSize: 64,
			want:    false,
		},
		{
			name:    "pass - Empty string",
			s:       "",
			base:    10,
			bitSize: 64,
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsStringNumericUint(tt.s, tt.base, tt.bitSize); got != tt.want {
				t.Errorf("IsStringNumericUint(%q) = %v, want %v", tt.s, got, tt.want)
			}
		})
	}
}

func TestIsMap(t *testing.T) {
	tests := []struct {
		name string
		m    interface{}
		want bool
	}{
		{
			name: "pass - Valid map",
			m:    map[string]interface{}{},
			want: true,
		}
		{
			name: "pass - Invalid map",
			m:    "Invalid map",
			want: false,
		}
		{
			name: "pass - Invalid map (2)",
			m:    int(2),
			want: false,
		}
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsMap(tt.m); got != tt.want {
				t.Errorf("IsMap(%v) = %v, want %v", tt.m, got, tt.want)
			}
		})
	}
}

func TestIsXRPLNumber(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  bool
	}{
		// Valid cases - integers
		{
			name:  "pass - Valid positive integer",
			value: "123",
			want:  true,
		},
		{
			name:  "pass - Valid negative integer",
			value: "-456",
			want:  true,
		},
		{
			name:  "pass - Valid integer with plus sign",
			value: "+789",
			want:  true,
		},
		{
			name:  "pass - Valid zero",
			value: "0",
			want:  true,
		},
		{
			name:  "pass - Valid integer with leading zeros",
			value: "007",
			want:  true,
		},
		// Valid cases - decimals
		{
			name:  "pass - Valid positive decimal",
			value: "123.456",
			want:  true,
		},
		{
			name:  "pass - Valid negative decimal",
			value: "-987.654",
			want:  true,
		},
		{
			name:  "pass - Valid decimal with plus sign",
			value: "+3.14",
			want:  true,
		},
		{
			name:  "pass - Valid decimal with trailing dot",
			value: "123.",
			want:  true,
		},
		{
			name:  "pass - Valid decimal starting with dot",
			value: ".5",
			want:  true,
		},
		{
			name:  "pass - Valid negative decimal starting with dot",
			value: "-.5",
			want:  true,
		},
		{
			name:  "pass - Valid positive decimal starting with dot",
			value: "+.5",
			want:  true,
		},
		{
			name:  "pass - Valid zero decimal",
			value: "0.0",
			want:  true,
		},
		// Valid cases - scientific notation
		{
			name:  "pass - Valid positive scientific notation lowercase e",
			value: "+3.14e10",
			want:  true,
		},
		{
			name:  "pass - Valid negative scientific notation lowercase e",
			value: "-7.2e-9",
			want:  true,
		},
		{
			name:  "pass - Valid scientific notation uppercase E",
			value: "123E5",
			want:  true,
		},
		{
			name:  "pass - Valid scientific notation with positive exponent",
			value: "123e+5",
			want:  true,
		},
		{
			name:  "pass - Valid scientific notation with negative exponent",
			value: "123e-5",
			want:  true,
		},
		{
			name:  "pass - Valid scientific notation integer base",
			value: "123e10",
			want:  true,
		},
		{
			name:  "pass - Valid scientific notation decimal base",
			value: "1.5e10",
			want:  true,
		},
		{
			name:  "pass - Valid scientific notation starting with dot",
			value: ".5e10",
			want:  true,
		},
		{
			name:  "pass - Valid scientific notation with trailing dot",
			value: "123.e10",
			want:  true,
		},
		// Valid cases - whitespace handling (should be trimmed)
		{
			name:  "pass - Valid number with leading whitespace",
			value: "  123",
			want:  true,
		},
		{
			name:  "pass - Valid number with trailing whitespace",
			value: "123  ",
			want:  true,
		},
		{
			name:  "pass - Valid number with both leading and trailing whitespace",
			value: "  123.456  ",
			want:  true,
		},
		{
			name:  "pass - Valid number with tab whitespace",
			value: "\t123\t",
			want:  true,
		},
		{
			name:  "pass - Valid number with newline whitespace",
			value: "\n123\n",
			want:  true,
		},
		// Invalid cases - non-string types
		{
			name:  "fail - Non-string type (int)",
			value: 123,
			want:  false,
		},
		{
			name:  "fail - Non-string type (float64)",
			value: 123.456,
			want:  false,
		},
		{
			name:  "fail - Non-string type (bool)",
			value: true,
			want:  false,
		},
		{
			name:  "fail - Non-string type (nil)",
			value: nil,
			want:  false,
		},
		// Invalid cases - empty strings
		{
			name:  "fail - Empty string",
			value: "",
			want:  false,
		},
		{
			name:  "fail - Whitespace only string",
			value: "   ",
			want:  false,
		},
		{
			name:  "fail - Tab only string",
			value: "\t",
			want:  false,
		},
		{
			name:  "fail - Newline only string",
			value: "\n",
			want:  false,
		},
		// Invalid cases - just signs
		{
			name:  "fail - Just plus sign",
			value: "+",
			want:  false,
		},
		{
			name:  "fail - Just minus sign",
			value: "-",
			want:  false,
		},
		{
			name:  "fail - Just plus sign with whitespace",
			value: "  +  ",
			want:  false,
		},
		// Invalid cases - invalid formats
		{
			name:  "fail - Multiple decimal points",
			value: "123.45.67",
			want:  false,
		},
		{
			name:  "fail - Non-numeric characters",
			value: "123abc",
			want:  false,
		},
		{
			name:  "fail - Non-numeric characters in decimal",
			value: "123.45abc",
			want:  false,
		},
		{
			name:  "fail - Invalid exponent format (missing digits)",
			value: "123e",
			want:  false,
		},
		{
			name:  "fail - Invalid exponent format (just e)",
			value: "e10",
			want:  false,
		},
		{
			name:  "fail - Invalid exponent format (double e)",
			value: "123ee10",
			want:  false,
		},
		{
			name:  "fail - Invalid exponent format (missing exponent digits)",
			value: "123e+",
			want:  false,
		},
		{
			name:  "fail - Invalid exponent format (missing exponent digits negative)",
			value: "123e-",
			want:  false,
		},
		{
			name:  "fail - Spaces in middle of number",
			value: "12 3",
			want:  false,
		},
		{
			name:  "fail - Spaces in middle of decimal",
			value: "12. 3",
			want:  false,
		},
		{
			name:  "fail - Spaces in middle of exponent",
			value: "123e 10",
			want:  false,
		},
		{
			name:  "fail - Special characters",
			value: "123!",
			want:  false,
		},
		{
			name:  "fail - Special characters in decimal",
			value: "123.45@",
			want:  false,
		},
		{
			name:  "fail - Invalid decimal format (just dot)",
			value: ".",
			want:  false,
		},
		{
			name:  "fail - Invalid decimal format (dot with sign only)",
			value: "+.",
			want:  false,
		},
		{
			name:  "fail - Invalid decimal format (dot with sign only negative)",
			value: "-.",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsXRPLNumber(tt.value); got != tt.want {
				t.Errorf("IsXRPLNumber(%v) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}
