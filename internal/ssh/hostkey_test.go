package ssh

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewHostKeyVerifier(t *testing.T) {
	tests := []struct {
		name           string
		strictChecking bool
		acceptNewKeys  bool
		knownHostsPath string
		expectError    bool
	}{
		{
			name:           "strict checking enabled",
			strictChecking: true,
			acceptNewKeys:  false,
			knownHostsPath: "",
		},
		{
			name:           "accept new keys enabled",
			strictChecking: false,
			acceptNewKeys:  true,
			knownHostsPath: "",
		},
		{
			name:           "custom known_hosts path",
			strictChecking: true,
			acceptNewKeys:  false,
			knownHostsPath: "/custom/path/known_hosts",
		},
		{
			name:           "default settings",
			strictChecking: false,
			acceptNewKeys:  false,
			knownHostsPath: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			verifier, err := NewHostKeyVerifier(tt.strictChecking, tt.acceptNewKeys, tt.knownHostsPath)

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

			if verifier == nil {
				t.Errorf("verifier is nil")
				return
			}

			if verifier.strictChecking != tt.strictChecking {
				t.Errorf("strictChecking = %v, want %v", verifier.strictChecking, tt.strictChecking)
			}

			if verifier.acceptNewKeys != tt.acceptNewKeys {
				t.Errorf("acceptNewKeys = %v, want %v", verifier.acceptNewKeys, tt.acceptNewKeys)
			}

			// If custom path was provided, it should be used
			if tt.knownHostsPath != "" && verifier.knownHostsPath != tt.knownHostsPath {
				t.Errorf("knownHostsPath = %v, want %v", verifier.knownHostsPath, tt.knownHostsPath)
			}

			// If no path provided, should default to ~/.ssh/known_hosts
			if tt.knownHostsPath == "" {
				homeDir, _ := os.UserHomeDir()
				expectedPath := filepath.Join(homeDir, ".ssh", "known_hosts")
				if verifier.knownHostsPath != expectedPath {
					t.Errorf("knownHostsPath = %v, want %v", verifier.knownHostsPath, expectedPath)
				}
			}
		})
	}
}

func TestCreateKnownHostsFile(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()
	knownHostsPath := filepath.Join(tmpDir, ".ssh", "known_hosts")

	verifier := &HostKeyVerifier{
		knownHostsPath: knownHostsPath,
	}

	err := verifier.createKnownHostsFile()
	if err != nil {
		t.Fatalf("createKnownHostsFile() error = %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(knownHostsPath); os.IsNotExist(err) {
		t.Errorf("known_hosts file was not created")
	}

	// Verify directory exists
	if _, err := os.Stat(filepath.Dir(knownHostsPath)); os.IsNotExist(err) {
		t.Errorf(".ssh directory was not created")
	}

	// Verify file permissions (should be 0600)
	fileInfo, err := os.Stat(knownHostsPath)
	if err != nil {
		t.Fatalf("failed to stat known_hosts: %v", err)
	}

	// On Unix systems, check permissions
	if fileInfo.Mode().Perm() != 0600 {
		t.Logf("Warning: file permissions = %v, expected 0600 (may differ on Windows)", fileInfo.Mode().Perm())
	}

	// Verify directory permissions (should be 0700)
	dirInfo, err := os.Stat(filepath.Dir(knownHostsPath))
	if err != nil {
		t.Fatalf("failed to stat directory: %v", err)
	}

	if dirInfo.Mode().Perm() != 0700 {
		t.Logf("Warning: directory permissions = %v, expected 0700 (may differ on Windows)", dirInfo.Mode().Perm())
	}
}

func TestCreateKnownHostsFile_AlreadyExists(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()
	sshDir := filepath.Join(tmpDir, ".ssh")
	knownHostsPath := filepath.Join(sshDir, "known_hosts")

	// Pre-create the directory and file
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		t.Fatalf("failed to create .ssh dir: %v", err)
	}

	if err := os.WriteFile(knownHostsPath, []byte("existing content\n"), 0600); err != nil {
		t.Fatalf("failed to create known_hosts: %v", err)
	}

	verifier := &HostKeyVerifier{
		knownHostsPath: knownHostsPath,
	}

	// Should not fail if file already exists
	err := verifier.createKnownHostsFile()
	if err != nil {
		t.Errorf("createKnownHostsFile() should not error on existing file, got: %v", err)
	}
}

func TestGetCallback_FileNotFound(t *testing.T) {
	// Use a non-existent path
	tmpDir := t.TempDir()
	knownHostsPath := filepath.Join(tmpDir, "nonexistent", "known_hosts")

	verifier := &HostKeyVerifier{
		knownHostsPath: knownHostsPath,
		strictChecking: true,
		acceptNewKeys:  false,
	}

	// Should fail because file doesn't exist and we're in strict mode
	_, err := verifier.GetCallback()
	if err == nil {
		t.Error("GetCallback() should fail when known_hosts doesn't exist in strict mode")
	}
}

func TestGetCallback_NonStrictCreatesFile(t *testing.T) {
	// Use a non-existent path
	tmpDir := t.TempDir()
	knownHostsPath := filepath.Join(tmpDir, ".ssh", "known_hosts")

	verifier := &HostKeyVerifier{
		knownHostsPath: knownHostsPath,
		strictChecking: false,
		acceptNewKeys:  false,
	}

	// Should create the file in non-strict mode
	callback, err := verifier.GetCallback()
	if err != nil {
		t.Errorf("GetCallback() error in non-strict mode: %v", err)
	}

	if callback == nil {
		t.Error("GetCallback() returned nil callback")
	}

	// Verify file was created
	if _, err := os.Stat(knownHostsPath); os.IsNotExist(err) {
		t.Error("GetCallback() should have created known_hosts file in non-strict mode")
	}
}
