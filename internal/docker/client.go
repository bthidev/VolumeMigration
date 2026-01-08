package docker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

var (
	// ErrDockerNotAccessible is returned when Docker is not accessible
	ErrDockerNotAccessible = errors.New("docker is not accessible (not installed or permission denied)")

	// ErrContainerNotFound is returned when a container doesn't exist
	ErrContainerNotFound = errors.New("container not found")
)

// ContainerInfo holds information about a Docker container
type ContainerInfo struct {
	ID     string
	Name   string
	Mounts []MountInfo
}

// MountInfo holds information about a container mount
type MountInfo struct {
	Type        string
	Name        string // Volume name (for named volumes)
	Source      string
	Destination string
}

// Client wraps Docker operations
type Client struct {
	sudo *SudoDetector
	ctx  context.Context
}

// NewClient creates a new Docker client
func NewClient(ctx context.Context) (*Client, error) {
	sudo := NewSudoDetector()

	// Detect sudo requirement
	if err := sudo.Detect(ctx); err != nil {
		return nil, err
	}

	return &Client{
		sudo: sudo,
		ctx:  ctx,
	}, nil
}

// InspectContainer retrieves detailed information about a container
func (c *Client) InspectContainer(name string) (*ContainerInfo, error) {
	cmd := c.sudo.WrapCommand(c.ctx, "inspect", name)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if strings.Contains(stderr.String(), "No such object") ||
			strings.Contains(stderr.String(), "Error: No such container") {
			return nil, ErrContainerNotFound
		}
		return nil, fmt.Errorf("failed to inspect container: %w", err)
	}

	// Parse JSON output
	var inspectData []struct {
		ID     string `json:"Id"`
		Name   string `json:"Name"`
		Mounts []struct {
			Type        string `json:"Type"`
			Name        string `json:"Name"`
			Source      string `json:"Source"`
			Destination string `json:"Destination"`
		} `json:"Mounts"`
	}

	if err := json.Unmarshal(stdout.Bytes(), &inspectData); err != nil {
		return nil, fmt.Errorf("failed to parse inspect output: %w", err)
	}

	if len(inspectData) == 0 {
		return nil, ErrContainerNotFound
	}

	data := inspectData[0]
	info := &ContainerInfo{
		ID:   data.ID,
		Name: strings.TrimPrefix(data.Name, "/"),
	}

	for _, m := range data.Mounts {
		info.Mounts = append(info.Mounts, MountInfo{
			Type:        m.Type,
			Name:        m.Name,
			Source:      m.Source,
			Destination: m.Destination,
		})
	}

	return info, nil
}

// ListVolumes returns a list of volume names used by a container
func (c *Client) ListVolumes(containerName string) ([]string, error) {
	info, err := c.InspectContainer(containerName)
	if err != nil {
		return nil, err
	}

	var volumes []string
	for _, mount := range info.Mounts {
		// Only include named volumes, skip bind mounts
		if mount.Type == "volume" && mount.Name != "" {
			volumes = append(volumes, mount.Name)
		}
	}

	return volumes, nil
}

// ValidateVolume checks if a volume exists
func (c *Client) ValidateVolume(volumeName string) error {
	cmd := c.sudo.WrapCommand(c.ctx, "volume", "inspect", volumeName)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if strings.Contains(stderr.String(), "No such volume") {
			return fmt.Errorf("volume %s not found", volumeName)
		}
		return fmt.Errorf("failed to inspect volume: %w", err)
	}

	return nil
}

// RequiresSudo returns whether Docker commands require sudo
func (c *Client) RequiresSudo() bool {
	return c.sudo.IsRequired()
}

// ExecCommand executes a Docker command and returns stdout
func (c *Client) ExecCommand(args ...string) (string, error) {
	cmd := c.sudo.WrapCommand(c.ctx, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("docker command failed: %w, stderr: %s", err, stderr.String())
	}

	return stdout.String(), nil
}

// ExecCommandWithOutput executes a Docker command and streams output
func (c *Client) ExecCommandWithOutput(stdout, stderr *bytes.Buffer, args ...string) error {
	cmd := c.sudo.WrapCommand(c.ctx, args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	return cmd.Run()
}
