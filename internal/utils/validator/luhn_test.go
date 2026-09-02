package validator

import (
	"testing"
)

func TestValidateLuhn(t *testing.T) {
	tests := []struct {
		name     string
		number   string
		expected bool
	}{
		{
			name:     "valid",
			number:   "12345678903",
			expected: true,
		},
		{
			name:     "invalid",
			number:   "12345678902",
			expected: false,
		},
		{
			name:     "minimal valid",
			number:   "0",
			expected: true,
		},
		{
			name:     "minimal invalid",
			number:   "1",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateLuhn(tt.number)
			if result != tt.expected {
				t.Errorf("ValidateLuhn(%q) = %v, expected %v", tt.number, result, tt.expected)
			}
		})
	}
}
