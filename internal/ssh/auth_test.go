package ssh

import (
	"os"
	"testing"
)

func TestParseHostPort(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantUser    string
		wantHost    string
		wantPort    string
		expectError bool
		setupEnv    func()
		cleanupEnv  func()
	}{
		{
			name:     "user@host",
			input:    "admin@example.com",
			wantUser: "admin",
			wantHost: "example.com",
			wantPort: "22",
		},
		{
			name:     "user@host:port",
			input:    "admin@example.com:2222",
			wantUser: "admin",
			wantHost: "example.com",
			wantPort: "2222",
		},
		{
			name:     "user@ip",
			input:    "root@192.168.1.100",
			wantUser: "root",
			wantHost: "192.168.1.100",
			wantPort: "22",
		},
		{
			name:     "user@ip:port",
			input:    "root@192.168.1.100:22",
			wantUser: "root",
			wantHost: "192.168.1.100",
			wantPort: "22",
		},
		{
			name:     "user@localhost",
			input:    "user@localhost",
			wantUser: "user",
			wantHost: "localhost",
			wantPort: "22",
		},
		{
			name:        "empty host",
			input:       "user@",
			expectError: true,
		},
		{
			name:        "no host",
			input:       "",
			expectError: true,
		},
		{
			name:     "host only with USER env",
			input:    "example.com",
			wantUser: "testuser",
			wantHost: "example.com",
			wantPort: "22",
			setupEnv: func() {
				os.Setenv("USER", "testuser")
			},
			cleanupEnv: func() {
				os.Unsetenv("USER")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupEnv != nil {
				tt.setupEnv()
			}
			if tt.cleanupEnv != nil {
				defer tt.cleanupEnv()
			}

			user, host, port, err := parseHostPort(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if user != tt.wantUser {
				t.Errorf("user = %v, want %v", user, tt.wantUser)
			}
			if host != tt.wantHost {
				t.Errorf("host = %v, want %v", host, tt.wantHost)
			}
			if port != tt.wantPort {
				t.Errorf("port = %v, want %v", port, tt.wantPort)
			}
		})
	}
}

func TestFindAt(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"user@host", 4},
		{"admin@example.com", 5},
		{"noatsign", -1},
		{"", -1},
		{"@start", 0},
		{"multiple@at@signs", 8}, // Returns first @
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := findAt(tt.input)
			if got != tt.want {
				t.Errorf("findAt(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestFindColon(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"host:22", 4},
		{"example.com:2222", 11},
		{"nocolon", -1},
		{"", -1},
		{":start", 0},
		{"multiple:colon:signs", 8}, // Returns first :
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := findColon(tt.input)
			if got != tt.want {
				t.Errorf("findColon(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidateKeyPermissions(t *testing.T) {
	// Create temporary directory for test files
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		permissions os.FileMode
		expectError bool
	}{
		{
			name:        "secure 0600 permissions",
			permissions: 0600,
			expectError: false,
		},
		{
			name:        "secure 0400 permissions",
			permissions: 0400,
			expectError: false,
		},
		{
			name:        "insecure 0644 permissions (world readable)",
			permissions: 0644,
			expectError: true,
		},
		{
			name:        "insecure 0666 permissions (world writable)",
			permissions: 0666,
			expectError: true,
		},
		{
			name:        "insecure 0640 permissions (group readable)",
			permissions: 0640,
			expectError: true,
		},
		{
			name:        "insecure 0777 permissions (all)",
			permissions: 0777,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test file with specific permissions
			testFile := tmpDir + "/" + tt.name
			err := os.WriteFile(testFile, []byte("test key content"), tt.permissions)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Validate permissions
			err = validateKeyPermissions(testFile)

			if tt.expectError && err == nil {
				t.Errorf("Expected error for permissions %o, but got none", tt.permissions)
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error for permissions %o, but got: %v", tt.permissions, err)
			}
		})
	}
}

func TestValidateKeyPermissions_NonExistentFile(t *testing.T) {
	err := validateKeyPermissions("/nonexistent/path/to/key")
	if err == nil {
		t.Error("Expected error for non-existent file, but got none")
	}
}
