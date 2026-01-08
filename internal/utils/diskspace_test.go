package utils

import (
	"testing"
)

func TestCalculateRequiredSpace(t *testing.T) {
	tests := []struct {
		name          string
		volumeSize    int64
		expectedSpace int64
	}{
		{
			name:          "1GB volume",
			volumeSize:    1073741824,      // 1 GB
			expectedSpace: 1181116006,      // 1.1 GB
		},
		{
			name:          "100MB volume",
			volumeSize:    104857600,       // 100 MB
			expectedSpace: 115343360,       // 110 MB
		},
		{
			name:          "10GB volume",
			volumeSize:    10737418240,     // 10 GB
			expectedSpace: 11811160064,     // 11 GB
		},
		{
			name:          "1MB volume",
			volumeSize:    1048576,         // 1 MB
			expectedSpace: 1153433,         // 1.1 MB
		},
		{
			name:          "zero volume",
			volumeSize:    0,
			expectedSpace: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateRequiredSpace(tt.volumeSize)
			if result != tt.expectedSpace {
				t.Errorf("CalculateRequiredSpace(%d) = %d, want %d", tt.volumeSize, result, tt.expectedSpace)
			}
		})
	}
}

func TestCalculateRequiredSpace_BufferPercentage(t *testing.T) {
	// Verify the 10% buffer is correctly applied
	volumeSize := int64(1000000) // 1 million bytes

	result := CalculateRequiredSpace(volumeSize)
	expected := int64(1100000) // 1.1 million bytes

	if result != expected {
		t.Errorf("CalculateRequiredSpace(%d) = %d, want %d (10%% buffer)", volumeSize, result, expected)
	}

	// Verify it's exactly 10% more
	buffer := result - volumeSize
	expectedBuffer := volumeSize / 10

	if buffer != expectedBuffer {
		t.Errorf("Buffer = %d, want %d (10%% of volume size)", buffer, expectedBuffer)
	}
}

func TestValidateDiskSpace_SufficientSpace(t *testing.T) {
	tests := []struct {
		name      string
		location  string
		required  uint64
		available uint64
	}{
		{
			name:      "exactly enough space",
			location:  "/tmp",
			required:  1024,
			available: 1024,
		},
		{
			name:      "more than enough space",
			location:  "/tmp",
			required:  1024,
			available: 2048,
		},
		{
			name:      "much more space",
			location:  "/var",
			required:  1024 * 1024,
			available: 10 * 1024 * 1024,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDiskSpace(tt.location, tt.required, tt.available)
			if err != nil {
				t.Errorf("ValidateDiskSpace() should not error with sufficient space, got: %v", err)
			}
		})
	}
}

func TestValidateDiskSpace_InsufficientSpace(t *testing.T) {
	tests := []struct {
		name      string
		location  string
		required  uint64
		available uint64
	}{
		{
			name:      "no space available",
			location:  "/tmp",
			required:  1024,
			available: 0,
		},
		{
			name:      "not enough space",
			location:  "/tmp",
			required:  2048,
			available: 1024,
		},
		{
			name:      "much less space",
			location:  "/var",
			required:  10 * 1024 * 1024,
			available: 1024 * 1024,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDiskSpace(tt.location, tt.required, tt.available)
			if err == nil {
				t.Errorf("ValidateDiskSpace() should error with insufficient space, got nil")
			}
			// Error message should contain "insufficient disk space"
			if err != nil && err.Error() == "" {
				t.Error("Error message should not be empty")
			}
		})
	}
}

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
			want:  "1023.9 MB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatBytes(tt.bytes)
			if got != tt.want {
				t.Errorf("FormatBytes(%d) = %v, want %v", tt.bytes, got, tt.want)
			}
		})
	}
}

func TestFormatBytes_Consistency(t *testing.T) {
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
		result := FormatBytes(tc.bytes)
		// Should contain the expected unit
		if !containsUnit(result, tc.unit) {
			t.Errorf("FormatBytes(%d) = %q, expected to contain unit %q", tc.bytes, result, tc.unit)
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
		b.Run(FormatBytes(size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				FormatBytes(size)
			}
		})
	}
}

func BenchmarkCalculateRequiredSpace(b *testing.B) {
	volumeSizes := []int64{
		100 * 1024 * 1024,          // 100 MB
		1024 * 1024 * 1024,          // 1 GB
		10 * 1024 * 1024 * 1024,     // 10 GB
		100 * 1024 * 1024 * 1024,    // 100 GB
	}

	for _, size := range volumeSizes {
		b.Run(FormatBytes(size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				CalculateRequiredSpace(size)
			}
		})
	}
}
