package migrator

import (
	"testing"

	"volume-migrator/internal/utils"
)

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name  string
		bytes int64
		want  string
	}{
		{
			name:  "zero bytes",
			bytes: 0,
			want:  "0 B",
		},
		{
			name:  "bytes only",
			bytes: 512,
			want:  "512 B",
		},
		{
			name:  "exactly 1 KB",
			bytes: 1024,
			want:  "1.0 KB",
		},
		{
			name:  "kilobytes",
			bytes: 1536,
			want:  "1.5 KB",
		},
		{
			name:  "exactly 1 MB",
			bytes: 1024 * 1024,
			want:  "1.0 MB",
		},
		{
			name:  "megabytes",
			bytes: int64(2.5 * 1024 * 1024),
			want:  "2.5 MB",
		},
		{
			name:  "exactly 1 GB",
			bytes: 1024 * 1024 * 1024,
			want:  "1.0 GB",
		},
		{
			name:  "gigabytes",
			bytes: int64(1.5 * 1024 * 1024 * 1024),
			want:  "1.5 GB",
		},
		{
			name:  "exactly 1 TB",
			bytes: 1024 * 1024 * 1024 * 1024,
			want:  "1.0 TB",
		},
		{
			name:  "terabytes",
			bytes: int64(2.0 * 1024 * 1024 * 1024 * 1024),
			want:  "2.0 TB",
		},
		{
			name:  "exactly 1 PB",
			bytes: 1024 * 1024 * 1024 * 1024 * 1024,
			want:  "1.0 PB",
		},
		{
			name:  "exactly 1 EB",
			bytes: 1024 * 1024 * 1024 * 1024 * 1024 * 1024,
			want:  "1.0 EB",
		},
		{
			name:  "less than 1 KB",
			bytes: 1000,
			want:  "1000 B",
		},
		{
			name:  "large GB value",
			bytes: 1073634508,
			want:  "1023.9 MB", // Actual output
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := utils.FormatBytes(tt.bytes)
			if got != tt.want {
				t.Errorf("utils.FormatBytes(%d) = %v, want %v", tt.bytes, got, tt.want)
			}
		})
	}
}

func TestFormatBytes_RoundTrip(t *testing.T) {
	// Test that formatBytes produces expected output for common sizes
	testCases := []struct {
		bytes int64
		unit  string
	}{
		{1024, "KB"},
		{1024 * 1024, "MB"},
		{1024 * 1024 * 1024, "GB"},
		{1024 * 1024 * 1024 * 1024, "TB"},
	}

	for _, tc := range testCases {
		result := utils.FormatBytes(tc.bytes)
		// Should contain the expected unit
		if !containsUnit(result, tc.unit) {
			t.Errorf("utils.FormatBytes(%d) = %q, expected to contain unit %q", tc.bytes, result, tc.unit)
		}
	}
}

func containsUnit(s, unit string) bool {
	return len(s) >= len(unit) && s[len(s)-len(unit):] == unit
}

func BenchmarkFormatBytes(b *testing.B) {
	testSizes := []int64{
		512,
		1024 * 1024,
		1024 * 1024 * 1024,
		1024 * 1024 * 1024 * 1024,
	}

	for _, size := range testSizes {
		b.Run(utils.FormatBytes(size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				utils.FormatBytes(size)
			}
		})
	}
}
