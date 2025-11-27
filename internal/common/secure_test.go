package common

import (
	"testing"
)

func TestGenerateSecureRandomString(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{"length 8", 8},
		{"length 16", 16},
		{"length 32", 32},
		{"length 64", 64},
		{"length 128", 128},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GenerateSecureRandomString(tt.length)
			if err != nil {
				t.Fatalf("GenerateSecureRandomString(%d) returned error: %v", tt.length, err)
			}
			if len(result) != tt.length {
				t.Errorf("GenerateSecureRandomString(%d) returned string of length %d, want %d", tt.length, len(result), tt.length)
			}
		})
	}
}

func TestGenerateSecureRandomString_Uniqueness(t *testing.T) {
	const iterations = 100
	const length = 32

	generated := make(map[string]bool)

	for i := 0; i < iterations; i++ {
		result, err := GenerateSecureRandomString(length)
		if err != nil {
			t.Fatalf("GenerateSecureRandomString(%d) returned error: %v", length, err)
		}
		if generated[result] {
			t.Errorf("GenerateSecureRandomString generated duplicate string: %s", result)
		}
		generated[result] = true
	}
}

func TestGenerateSecureRandomString_ValidCharacters(t *testing.T) {
	result, err := GenerateSecureRandomString(100)
	if err != nil {
		t.Fatalf("GenerateSecureRandomString(100) returned error: %v", err)
	}

	// URL-safe base64 characters: A-Z, a-z, 0-9, -, _
	for i, c := range result {
		valid := (c >= 'A' && c <= 'Z') ||
			(c >= 'a' && c <= 'z') ||
			(c >= '0' && c <= '9') ||
			c == '-' || c == '_'
		if !valid {
			t.Errorf("Invalid character '%c' at position %d", c, i)
		}
	}
}

func TestGenerateSecureRandomString_ZeroLength(t *testing.T) {
	result, err := GenerateSecureRandomString(0)
	if err != nil {
		t.Fatalf("GenerateSecureRandomString(0) returned error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("GenerateSecureRandomString(0) returned string of length %d, want 0", len(result))
	}
}

func BenchmarkGenerateSecureRandomString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = GenerateSecureRandomString(32)
	}
}
