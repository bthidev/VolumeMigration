package docker

import (
	"testing"
)

func TestParseSizeToBytes(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int64
	}{
		{
			name:  "bytes",
			input: "100B",
			want:  100,
		},
		{
			name:  "kilobytes",
			input: "1KB",
			want:  1024,
		},
		{
			name:  "kilobytes with decimal",
			input: "1.5KB",
			want:  1536,
		},
		{
			name:  "megabytes",
			input: "10MB",
			want:  10 * 1024 * 1024,
		},
		{
			name:  "megabytes with decimal",
			input: "2.5MB",
			want:  int64(2.5 * 1024 * 1024),
		},
		{
			name:  "gigabytes",
			input: "1GB",
			want:  1024 * 1024 * 1024,
		},
		{
			name:  "gigabytes with decimal",
			input: "1.2GB",
			want:  1288490188,
		},
		{
			name:  "terabytes",
			input: "1TB",
			want:  1024 * 1024 * 1024 * 1024,
		},
		{
			name:  "terabytes with decimal",
			input: "1.5TB",
			want:  int64(1.5 * 1024 * 1024 * 1024 * 1024),
		},
		{
			name:  "zero bytes",
			input: "0B",
			want:  0,
		},
		{
			name:  "invalid format",
			input: "invalid",
			want:  0,
		},
		{
			name:  "empty string",
			input: "",
			want:  0,
		},
		{
			name:  "lowercase units",
			input: "100kb",
			want:  0, // parseSizeToBytes expects uppercase, lowercase doesn't match
		},
		{
			name:  "mixed case units",
			input: "5Mb",
			want:  0, // parseSizeToBytes expects uppercase, mixed case doesn't match
		},
		{
			name:  "whitespace",
			input: " 10MB ",
			want:  10 * 1024 * 1024,
		},
		{
			name:  "just number",
			input: "512",
			want:  512,
		},
		{
			name:  "decimal without unit",
			input: "1.5",
			want:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseSizeToBytes(tt.input)
			if got != tt.want {
				t.Errorf("parseSizeToBytes(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseSizeToBytes_EdgeCases(t *testing.T) {
	// Test very large numbers
	largeGB := "999GB"
	result := parseSizeToBytes(largeGB)
	expected := int64(999 * 1024 * 1024 * 1024)
	if result != expected {
		t.Errorf("parseSizeToBytes(%q) = %v, want %v", largeGB, result, expected)
	}

	// Test very small decimals
	smallMB := "0.001MB"
	result = parseSizeToBytes(smallMB)
	expected = int64(1048)
	if result != expected {
		t.Errorf("parseSizeToBytes(%q) = %v, want %v", smallMB, result, expected)
	}
}

func TestParseSizeToBytes_Consistency(t *testing.T) {
	// Verify that same values with different units are consistent
	testCases := []struct {
		input1 string
		input2 string
	}{
		{"1024B", "1KB"},
		{"1024KB", "1MB"},
		{"1024MB", "1GB"},
		{"1024GB", "1TB"},
	}

	for _, tc := range testCases {
		result1 := parseSizeToBytes(tc.input1)
		result2 := parseSizeToBytes(tc.input2)
		if result1 != result2 {
			t.Errorf("parseSizeToBytes(%q) = %v, parseSizeToBytes(%q) = %v, expected equal",
				tc.input1, result1, tc.input2, result2)
		}
	}
}

func BenchmarkParseSizeToBytes(b *testing.B) {
	testCases := []string{
		"100B",
		"1KB",
		"1.5MB",
		"10GB",
		"1.2TB",
	}

	for _, tc := range testCases {
		b.Run(tc, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				parseSizeToBytes(tc)
			}
		})
	}
}
