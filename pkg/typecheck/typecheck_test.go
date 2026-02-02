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