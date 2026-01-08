package docker

import (
	"context"
	"testing"
)

func TestClient_RequiresSudo(t *testing.T) {
	tests := []struct {
		name         string
		requiresSudo bool
		expected     bool
	}{
		{
			name:         "requires sudo",
			requiresSudo: true,
			expected:     true,
		},
		{
			name:         "does not require sudo",
			requiresSudo: false,
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := &SudoDetector{
				required: tt.requiresSudo,
			}
			client := &Client{
				sudo: detector,
			}

			if client.RequiresSudo() != tt.expected {
				t.Errorf("RequiresSudo() = %v, want %v", client.RequiresSudo(), tt.expected)
			}
		})
	}
}

func TestSudoDetector_IsRequired(t *testing.T) {
	tests := []struct {
		name     string
		required bool
		expected bool
	}{
		{
			name:     "sudo required",
			required: true,
			expected: true,
		},
		{
			name:     "sudo not required",
			required: false,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := &SudoDetector{
				required: tt.required,
			}

			if detector.IsRequired() != tt.expected {
				t.Errorf("IsRequired() = %v, want %v", detector.IsRequired(), tt.expected)
			}
		})
	}
}

func TestSudoDetector_WrapCommand(t *testing.T) {
	tests := []struct {
		name     string
		required bool
		args     []string
		wantCmd  string
	}{
		{
			name:     "without sudo",
			required: false,
			args:     []string{"ps", "-a"},
			wantCmd:  "docker",
		},
		{
			name:     "with sudo",
			required: true,
			args:     []string{"ps", "-a"},
			wantCmd:  "sudo",
		},
		{
			name:     "volume command without sudo",
			required: false,
			args:     []string{"volume", "ls"},
			wantCmd:  "docker",
		},
		{
			name:     "volume command with sudo",
			required: true,
			args:     []string{"volume", "ls"},
			wantCmd:  "sudo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := &SudoDetector{
				required: tt.required,
			}

			ctx := context.Background()
			cmd := detector.WrapCommand(ctx, tt.args...)

			if cmd.Path != tt.wantCmd && cmd.Args[0] != tt.wantCmd {
				t.Errorf("WrapCommand() command = %v, want %v", cmd.Args[0], tt.wantCmd)
			}

			// Verify args are properly included
			if tt.required {
				// With sudo: sudo docker <args>
				if len(cmd.Args) < 2 {
					t.Error("Expected sudo command to have docker in args")
				}
			} else {
				// Without sudo: docker <args>
				if len(cmd.Args) < len(tt.args) {
					t.Error("Expected docker command to have all args")
				}
			}
		})
	}
}

func TestNewSudoDetector(t *testing.T) {
	detector := NewSudoDetector()

	if detector == nil {
		t.Fatal("NewSudoDetector() returned nil")
	}

	// Should start with required = false (will be detected later)
	if detector.required {
		t.Error("New SudoDetector should start with required = false")
	}
}

func TestVolumeInfo_Structure(t *testing.T) {
	// Test that VolumeInfo struct can be created and fields accessed
	info := VolumeInfo{
		Name:       "test-volume",
		Container:  "test-container",
		MountPath:  "/data",
		Size:       "1.5 GB",
		SizeBytes:  1610612736,
		Selected:   true,
	}

	if info.Name != "test-volume" {
		t.Errorf("Name = %v, want %v", info.Name, "test-volume")
	}
	if info.Container != "test-container" {
		t.Errorf("Container = %v, want %v", info.Container, "test-container")
	}
	if info.MountPath != "/data" {
		t.Errorf("MountPath = %v, want %v", info.MountPath, "/data")
	}
	if info.Size != "1.5 GB" {
		t.Errorf("Size = %v, want %v", info.Size, "1.5 GB")
	}
	if info.SizeBytes != 1610612736 {
		t.Errorf("SizeBytes = %v, want %v", info.SizeBytes, 1610612736)
	}
	if !info.Selected {
		t.Error("Selected should be true")
	}
}

func TestContainerInfo_Structure(t *testing.T) {
	// Test that ContainerInfo struct can be created and fields accessed
	mount := MountInfo{
		Type:        "volume",
		Name:        "my-volume",
		Source:      "/var/lib/docker/volumes/my-volume/_data",
		Destination: "/app/data",
	}

	info := ContainerInfo{
		Name:   "my-container",
		ID:     "abc123",
		Mounts: []MountInfo{mount},
	}

	if info.Name != "my-container" {
		t.Errorf("Name = %v, want %v", info.Name, "my-container")
	}
	if info.ID != "abc123" {
		t.Errorf("ID = %v, want %v", info.ID, "abc123")
	}
	if len(info.Mounts) != 1 {
		t.Errorf("Mounts length = %v, want %v", len(info.Mounts), 1)
	}
	if info.Mounts[0].Name != "my-volume" {
		t.Errorf("Mount name = %v, want %v", info.Mounts[0].Name, "my-volume")
	}
}

func TestMountInfo_Structure(t *testing.T) {
	mount := MountInfo{
		Type:        "volume",
		Name:        "data-volume",
		Source:      "/var/lib/docker/volumes/data-volume/_data",
		Destination: "/data",
	}

	if mount.Type != "volume" {
		t.Errorf("Type = %v, want %v", mount.Type, "volume")
	}
	if mount.Name != "data-volume" {
		t.Errorf("Name = %v, want %v", mount.Name, "data-volume")
	}
	if mount.Destination != "/data" {
		t.Errorf("Destination = %v, want %v", mount.Destination, "/data")
	}
	if mount.Source != "/var/lib/docker/volumes/data-volume/_data" {
		t.Errorf("Source = %v, want %v", mount.Source, "/var/lib/docker/volumes/data-volume/_data")
	}
}
