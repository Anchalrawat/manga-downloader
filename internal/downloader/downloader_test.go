package downloader

import (
	"testing"
)

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Normal Name", "Normal Name"},
		{"Name: With Colons", "Name_ With Colons"},
		{"Name/With/Slashes", "Name_With_Slashes"},
		{"<Invalid>Chars?", "_Invalid_Chars_"},
		{"mixed|separators\\here", "mixed_separators_here"},
	}

	for _, tt := range tests {
		got := SanitizeFilename(tt.input)
		if got != tt.expected {
			t.Errorf("SanitizeFilename(%q) = %q; want %q", tt.input, got, tt.expected)
		}
	}
}
