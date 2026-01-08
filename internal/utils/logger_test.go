package utils

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func TestGetLogger(t *testing.T) {
	logger := GetLogger()
	if logger == nil {
		t.Fatal("GetLogger() returned nil")
	}

	// Verify it's a logrus logger (it should be a concrete type, not interface)
	// Type assertion not needed since GetLogger() returns *logrus.Logger directly

	// Verify default level is InfoLevel
	if logger.Level != logrus.InfoLevel {
		t.Errorf("default level = %v, want %v", logger.Level, logrus.InfoLevel)
	}
}

func TestGetLogger_Singleton(t *testing.T) {
	// Verify that GetLogger returns the same instance
	logger1 := GetLogger()
	logger2 := GetLogger()

	if logger1 != logger2 {
		t.Error("GetLogger() should return the same instance (singleton pattern)")
	}
}

func TestSetVerbose(t *testing.T) {
	tests := []struct {
		name    string
		verbose bool
		want    logrus.Level
	}{
		{
			name:    "verbose enabled",
			verbose: true,
			want:    logrus.DebugLevel,
		},
		{
			name:    "verbose disabled",
			verbose: false,
			want:    logrus.InfoLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetVerbose(tt.verbose)
			logger := GetLogger()

			if logger.Level != tt.want {
				t.Errorf("after SetVerbose(%v), level = %v, want %v", tt.verbose, logger.Level, tt.want)
			}
		})
	}
}

func TestSetVerbose_Toggle(t *testing.T) {
	// Test toggling verbose on and off
	logger := GetLogger()

	// Start with non-verbose
	SetVerbose(false)
	if logger.Level != logrus.InfoLevel {
		t.Errorf("initial level = %v, want InfoLevel", logger.Level)
	}

	// Enable verbose
	SetVerbose(true)
	if logger.Level != logrus.DebugLevel {
		t.Errorf("after SetVerbose(true), level = %v, want DebugLevel", logger.Level)
	}

	// Disable verbose
	SetVerbose(false)
	if logger.Level != logrus.InfoLevel {
		t.Errorf("after SetVerbose(false), level = %v, want InfoLevel", logger.Level)
	}

	// Enable again
	SetVerbose(true)
	if logger.Level != logrus.DebugLevel {
		t.Errorf("after second SetVerbose(true), level = %v, want DebugLevel", logger.Level)
	}
}

func TestLogger_Configuration(t *testing.T) {
	logger := GetLogger()

	// Verify formatter is TextFormatter
	if _, ok := logger.Formatter.(*logrus.TextFormatter); !ok {
		t.Error("logger formatter is not TextFormatter")
	}

	// Verify output is set (should be os.Stdout)
	if logger.Out == nil {
		t.Error("logger output is nil")
	}
}

func BenchmarkGetLogger(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetLogger()
	}
}

func BenchmarkSetVerbose(b *testing.B) {
	for i := 0; i < b.N; i++ {
		SetVerbose(i%2 == 0)
	}
}
