package docker

import (
	"context"
	"os/exec"
	"sync"
)

// SudoDetector handles detection and caching of sudo requirement for Docker commands
type SudoDetector struct {
	required bool
	checked  bool
	mu       sync.Mutex
}

// NewSudoDetector creates a new sudo detector
func NewSudoDetector() *SudoDetector {
	return &SudoDetector{}
}

// Detect checks if sudo is required to run Docker commands
// It tries to run "docker ps" without sudo first, then with sudo if that fails
func (sd *SudoDetector) Detect(ctx context.Context) error {
	sd.mu.Lock()
	defer sd.mu.Unlock()

	if sd.checked {
		return nil
	}

	// Try without sudo first
	cmd := exec.CommandContext(ctx, "docker", "ps")
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Run(); err == nil {
		// Docker works without sudo
		sd.required = false
		sd.checked = true
		return nil
	}

	// Try with sudo -n (non-interactive)
	cmd = exec.CommandContext(ctx, "sudo", "-n", "docker", "ps")
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return ErrDockerNotAccessible
	}

	sd.required = true
	sd.checked = true
	return nil
}

// IsRequired returns whether sudo is required
func (sd *SudoDetector) IsRequired() bool {
	sd.mu.Lock()
	defer sd.mu.Unlock()
	return sd.required
}

// WrapCommand wraps a docker command with sudo if required
func (sd *SudoDetector) WrapCommand(ctx context.Context, args ...string) *exec.Cmd {
	sd.mu.Lock()
	defer sd.mu.Unlock()

	if sd.required {
		sudoArgs := append([]string{"docker"}, args...)
		return exec.CommandContext(ctx, "sudo", sudoArgs...)
	}
	return exec.CommandContext(ctx, "docker", args...)
}
