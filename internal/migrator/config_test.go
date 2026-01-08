package migrator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateConfig_EmptyContainers(t *testing.T) {
	config := &Config{
		Containers: []string{},
		RemoteHost: "user@host",
	}

	err := ValidateConfig(config)
	if err == nil {
		t.Error("Expected error for empty containers, got nil")
	}
	if !strings.Contains(err.Error(), "no containers specified") {
		t.Errorf("Expected 'no containers specified' error, got: %v", err)
	}
}

func TestValidateConfig_EmptyContainerName(t *testing.T) {
	config := &Config{
		Containers: []string{"container1", "", "container2"},
		RemoteHost: "user@host",
	}

	err := ValidateConfig(config)
	if err == nil {
		t.Error("Expected error for empty container name, got nil")
	}
	if !strings.Contains(err.Error(), "container at index 1 is empty") {
		t.Errorf("Expected 'container at index 1 is empty' error, got: %v", err)
	}
}

func TestValidateConfig_WhitespaceContainerName(t *testing.T) {
	config := &Config{
		Containers: []string{"container1", "   ", "container2"},
		RemoteHost: "user@host",
	}

	err := ValidateConfig(config)
	if err == nil {
		t.Error("Expected error for whitespace container name, got nil")
	}
	if !strings.Contains(err.Error(), "container at index 1 is empty") {
		t.Errorf("Expected 'container at index 1 is empty' error, got: %v", err)
	}
}

func TestValidateConfig_EmptyRemoteHost(t *testing.T) {
	config := &Config{
		Containers: []string{"container1"},
		RemoteHost: "",
	}

	err := ValidateConfig(config)
	if err == nil {
		t.Error("Expected error for empty remote host, got nil")
	}
	if !strings.Contains(err.Error(), "remote host not specified") {
		t.Errorf("Expected 'remote host not specified' error, got: %v", err)
	}
}

func TestValidateConfig_InvalidRemoteHostFormat(t *testing.T) {
	tests := []struct {
		name       string
		remoteHost string
		errorPart  string
	}{
		{
			name:       "no @ symbol",
			remoteHost: "hostname",
			errorPart:  "must be in format 'user@host'",
		},
		{
			name:       "multiple @ symbols",
			remoteHost: "user@host@extra",
			errorPart:  "invalid remote host format",
		},
		{
			name:       "empty username",
			remoteHost: "@hostname",
			errorPart:  "username cannot be empty",
		},
		{
			name:       "empty host",
			remoteHost: "user@",
			errorPart:  "host cannot be empty",
		},
		{
			name:       "whitespace username",
			remoteHost: "   @hostname",
			errorPart:  "username cannot be empty",
		},
		{
			name:       "whitespace host",
			remoteHost: "user@   ",
			errorPart:  "host cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Containers: []string{"container1"},
				RemoteHost: tt.remoteHost,
			}

			err := ValidateConfig(config)
			if err == nil {
				t.Errorf("Expected error for %s, got nil", tt.name)
			}
			if !strings.Contains(err.Error(), tt.errorPart) {
				t.Errorf("Expected error containing '%s', got: %v", tt.errorPart, err)
			}
		})
	}
}

func TestValidateConfig_ValidRemoteHost(t *testing.T) {
	tests := []struct {
		name       string
		remoteHost string
	}{
		{
			name:       "user@host",
			remoteHost: "user@host",
		},
		{
			name:       "user@host:port",
			remoteHost: "user@192.168.1.100",
		},
		{
			name:       "user@domain",
			remoteHost: "admin@example.com",
		},
		{
			name:       "user@host with port in SSHPort",
			remoteHost: "user@host",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Containers: []string{"container1"},
				RemoteHost: tt.remoteHost,
			}

			err := ValidateConfig(config)
			if err != nil {
				t.Errorf("Expected no error for valid remote host, got: %v", err)
			}
		})
	}
}

func TestValidateConfig_InvalidSSHPort(t *testing.T) {
	tests := []struct {
		name     string
		port     string
		errorPart string
	}{
		{
			name:      "non-numeric port",
			port:      "abc",
			errorPart: "must be a number",
		},
		{
			name:      "port zero",
			port:      "0",
			errorPart: "must be between 1 and 65535",
		},
		{
			name:      "negative port",
			port:      "-1",
			errorPart: "must be between 1 and 65535",
		},
		{
			name:      "port too large",
			port:      "65536",
			errorPart: "must be between 1 and 65535",
		},
		{
			name:      "port way too large",
			port:      "999999",
			errorPart: "must be between 1 and 65535",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Containers: []string{"container1"},
				RemoteHost: "user@host",
				SSHPort:    tt.port,
			}

			err := ValidateConfig(config)
			if err == nil {
				t.Errorf("Expected error for %s, got nil", tt.name)
			}
			if !strings.Contains(err.Error(), tt.errorPart) {
				t.Errorf("Expected error containing '%s', got: %v", tt.errorPart, err)
			}
		})
	}
}

func TestValidateConfig_ValidSSHPort(t *testing.T) {
	tests := []struct {
		name string
		port string
	}{
		{"port 22", "22"},
		{"port 2222", "2222"},
		{"port 1", "1"},
		{"port 65535", "65535"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Containers: []string{"container1"},
				RemoteHost: "user@host",
				SSHPort:    tt.port,
			}

			err := ValidateConfig(config)
			if err != nil {
				t.Errorf("Expected no error for valid port %s, got: %v", tt.port, err)
			}
		})
	}
}

func TestValidateConfig_RelativeTempDirectory(t *testing.T) {
	config := &Config{
		Containers: []string{"container1"},
		RemoteHost: "user@host",
		TempDir:    "relative/path",
	}

	err := ValidateConfig(config)
	if err == nil {
		t.Error("Expected error for relative temp directory, got nil")
	}
	if !strings.Contains(err.Error(), "temp directory must be an absolute path") {
		t.Errorf("Expected 'temp directory must be an absolute path' error, got: %v", err)
	}
}

func TestValidateConfig_RelativeRemoteTempDirectory(t *testing.T) {
	config := &Config{
		Containers:    []string{"container1"},
		RemoteHost:    "user@host",
		RemoteTempDir: "relative/path",
	}

	err := ValidateConfig(config)
	if err == nil {
		t.Error("Expected error for relative remote temp directory, got nil")
	}
	if !strings.Contains(err.Error(), "remote temp directory must be an absolute path") {
		t.Errorf("Expected 'remote temp directory must be an absolute path' error, got: %v", err)
	}
}

func TestValidateConfig_AbsoluteTempDirectories(t *testing.T) {
	config := &Config{
		Containers:    []string{"container1"},
		RemoteHost:    "user@host",
		TempDir:       "/tmp/local",
		RemoteTempDir: "/tmp/remote",
	}

	err := ValidateConfig(config)
	if err != nil {
		t.Errorf("Expected no error for absolute temp directories, got: %v", err)
	}
}

func TestValidateConfig_ConflictingFlags(t *testing.T) {
	config := &Config{
		Containers:            []string{"container1"},
		RemoteHost:            "user@host",
		StrictHostKeyChecking: true,
		AcceptHostKey:         true,
	}

	err := ValidateConfig(config)
	if err == nil {
		t.Error("Expected error for conflicting flags, got nil")
	}
	if !strings.Contains(err.Error(), "conflicting flags") {
		t.Errorf("Expected 'conflicting flags' error, got: %v", err)
	}
}

func TestValidateConfig_NonConflictingFlags(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
	}{
		{
			name: "both false",
			config: &Config{
				Containers:            []string{"container1"},
				RemoteHost:            "user@host",
				StrictHostKeyChecking: false,
				AcceptHostKey:         false,
			},
		},
		{
			name: "only strict checking",
			config: &Config{
				Containers:            []string{"container1"},
				RemoteHost:            "user@host",
				StrictHostKeyChecking: true,
				AcceptHostKey:         false,
			},
		},
		{
			name: "only accept host key",
			config: &Config{
				Containers:            []string{"container1"},
				RemoteHost:            "user@host",
				StrictHostKeyChecking: false,
				AcceptHostKey:         true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config)
			if err != nil {
				t.Errorf("Expected no error for non-conflicting flags, got: %v", err)
			}
		})
	}
}

func TestValidateConfig_NonExistentSSHKeyPath(t *testing.T) {
	config := &Config{
		Containers: []string{"container1"},
		RemoteHost: "user@host",
		SSHKeyPath: "/nonexistent/path/to/key",
	}

	err := ValidateConfig(config)
	if err == nil {
		t.Error("Expected error for non-existent SSH key, got nil")
	}
	if !strings.Contains(err.Error(), "SSH key file does not exist") {
		t.Errorf("Expected 'SSH key file does not exist' error, got: %v", err)
	}
}

func TestValidateConfig_ExistingSSHKeyPath(t *testing.T) {
	// Create temporary SSH key file
	tempDir := t.TempDir()
	keyPath := filepath.Join(tempDir, "test_key")
	if err := os.WriteFile(keyPath, []byte("test key content"), 0600); err != nil {
		t.Fatalf("Failed to create test key file: %v", err)
	}

	config := &Config{
		Containers: []string{"container1"},
		RemoteHost: "user@host",
		SSHKeyPath: keyPath,
	}

	err := ValidateConfig(config)
	if err != nil {
		t.Errorf("Expected no error for existing SSH key path, got: %v", err)
	}
}

func TestValidateConfig_NonExistentKnownHostsWithStrictChecking(t *testing.T) {
	config := &Config{
		Containers:            []string{"container1"},
		RemoteHost:            "user@host",
		StrictHostKeyChecking: true,
		KnownHostsFile:        "/nonexistent/known_hosts",
	}

	err := ValidateConfig(config)
	if err == nil {
		t.Error("Expected error for non-existent known_hosts with strict checking, got nil")
	}
	if !strings.Contains(err.Error(), "known_hosts file does not exist") {
		t.Errorf("Expected 'known_hosts file does not exist' error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "--accept-host-key") {
		t.Error("Expected error to suggest --accept-host-key")
	}
}

func TestValidateConfig_NonExistentKnownHostsWithoutStrictChecking(t *testing.T) {
	config := &Config{
		Containers:            []string{"container1"},
		RemoteHost:            "user@host",
		StrictHostKeyChecking: false,
		KnownHostsFile:        "/nonexistent/known_hosts",
	}

	// Should not error when strict checking is disabled
	err := ValidateConfig(config)
	if err != nil {
		t.Errorf("Expected no error for non-existent known_hosts without strict checking, got: %v", err)
	}
}

func TestValidateConfig_ExistingKnownHostsWithStrictChecking(t *testing.T) {
	// Create temporary known_hosts file
	tempDir := t.TempDir()
	knownHostsPath := filepath.Join(tempDir, "known_hosts")
	if err := os.WriteFile(knownHostsPath, []byte("known hosts content"), 0644); err != nil {
		t.Fatalf("Failed to create test known_hosts file: %v", err)
	}

	config := &Config{
		Containers:            []string{"container1"},
		RemoteHost:            "user@host",
		StrictHostKeyChecking: true,
		KnownHostsFile:        knownHostsPath,
	}

	err := ValidateConfig(config)
	if err != nil {
		t.Errorf("Expected no error for existing known_hosts with strict checking, got: %v", err)
	}
}

func TestValidateConfig_CompleteValidConfig(t *testing.T) {
	// Create temporary files
	tempDir := t.TempDir()
	keyPath := filepath.Join(tempDir, "test_key")
	knownHostsPath := filepath.Join(tempDir, "known_hosts")

	if err := os.WriteFile(keyPath, []byte("test key"), 0600); err != nil {
		t.Fatalf("Failed to create test key: %v", err)
	}
	if err := os.WriteFile(knownHostsPath, []byte("known hosts"), 0644); err != nil {
		t.Fatalf("Failed to create known_hosts: %v", err)
	}

	config := &Config{
		Containers:            []string{"container1", "container2"},
		RemoteHost:            "admin@example.com",
		SSHPort:               "2222",
		TempDir:               "/tmp/local",
		RemoteTempDir:         "/tmp/remote",
		SSHKeyPath:            keyPath,
		StrictHostKeyChecking: true,
		KnownHostsFile:        knownHostsPath,
		Verbose:               true,
		NoCleanup:             false,
		Force:                 false,
	}

	err := ValidateConfig(config)
	if err != nil {
		t.Errorf("Expected no error for complete valid config, got: %v", err)
	}
}
