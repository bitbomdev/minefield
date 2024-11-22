package utils

import (
	"testing"
)

func TestUint32ToStr(t *testing.T) {
	tests := []struct {
		name     string
		input    uint32
		expected string
	}{
		{"zero value", 0, "0"},
		{"small number", 42, "42"},
		{"max uint32", 4294967295, "4294967295"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Uint32ToStr(tt.input)
			if result != tt.expected {
				t.Errorf("Uint32ToStr(%d) = %s; want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestStrToUint32(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    uint32
		expectError bool
	}{
		{"zero value", "0", 0, false},
		{"small number", "42", 42, false},
		{"max uint32", "4294967295", 4294967295, false},
		{"negative number", "-1", 0, true},
		{"invalid input", "abc", 0, true},
		{"overflow", "4294967296", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := StrToUint32(tt.input)
			if tt.expectError && err == nil {
				t.Errorf("StrToUint32(%s) expected error but got none", tt.input)
			}
			if !tt.expectError && err != nil {
				t.Errorf("StrToUint32(%s) unexpected error: %v", tt.input, err)
			}
			if !tt.expectError && result != tt.expected {
				t.Errorf("StrToUint32(%s) = %d; want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIntToUint32(t *testing.T) {
	tests := []struct {
		name        string
		input       int
		expected    uint32
		expectError bool
	}{
		{"zero value", 0, 0, false},
		{"small number", 42, 42, false},
		{"max uint32", 4294967295, 4294967295, false},
		{"negative number", -1, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := IntToUint32(tt.input)
			if tt.expectError && err == nil {
				t.Errorf("IntToUint32(%d) expected error but got none", tt.input)
			}
			if !tt.expectError && err != nil {
				t.Errorf("IntToUint32(%d) unexpected error: %v", tt.input, err)
			}
			if !tt.expectError && result != tt.expected {
				t.Errorf("IntToUint32(%d) = %d; want %d", tt.input, result, tt.expected)
			}
		})
	}
}
