package docker

import (
	"fmt"
	"regexp"
	"strings"
)

// sizeRegex is compiled once at package initialization for performance
var sizeRegex = regexp.MustCompile(`^([\d.]+)([KMGT]?B?)$`)

// VolumeInfo holds detailed information about a Docker volume
type VolumeInfo struct {
	Name       string
	Container  string
	MountPath  string
	Size       string
	SizeBytes  int64
	Selected   bool
}

// GetVolumeSize retrieves the size of a Docker volume
// Uses "docker system df -v" to get volume sizes
func (c *Client) GetVolumeSize(volumeName string) (string, int64, error) {
	output, err := c.ExecCommand("system", "df", "-v")
	if err != nil {
		return "", 0, fmt.Errorf("failed to get volume size: %w", err)
	}

	// Parse the output to find the volume
	lines := strings.Split(output, "\n")

	// Find the VOLUMES section
	inVolumesSection := false
	for _, line := range lines {
		if strings.Contains(line, "VOLUME NAME") {
			inVolumesSection = true
			continue
		}

		if inVolumesSection {
			// Check if this line contains our volume
			if strings.Contains(line, volumeName) {
				// Parse the size from the line
				// Format: VOLUME NAME    LINKS     SIZE
				fields := strings.Fields(line)
				if len(fields) >= 3 {
					sizeStr := fields[2]
					sizeBytes := parseSizeToBytes(sizeStr)
					return sizeStr, sizeBytes, nil
				}
			}
		}
	}

	// If not found in df output, return 0
	return "0B", 0, nil
}

// GetVolumeMountPoints retrieves mount point information for a volume
func (c *Client) GetVolumeMountPoints(containerName, volumeName string) (string, error) {
	info, err := c.InspectContainer(containerName)
	if err != nil {
		return "", err
	}

	for _, mount := range info.Mounts {
		if mount.Type == "volume" && mount.Name == volumeName {
			return mount.Destination, nil
		}
	}

	return "", fmt.Errorf("volume %s not found in container %s", volumeName, containerName)
}

// GetAllVolumesInfo retrieves detailed information about all volumes from specified containers
func (c *Client) GetAllVolumesInfo(containerNames []string) ([]VolumeInfo, error) {
	volumeMap := make(map[string]*VolumeInfo) // deduplicate volumes

	for _, containerName := range containerNames {
		volumes, err := c.ListVolumes(containerName)
		if err != nil {
			return nil, fmt.Errorf("failed to list volumes for container %s: %w", containerName, err)
		}

		for _, volumeName := range volumes {
			// Skip if already processed
			if _, exists := volumeMap[volumeName]; exists {
				continue
			}

			mountPath, err := c.GetVolumeMountPoints(containerName, volumeName)
			if err != nil {
				mountPath = "N/A"
			}

			size, sizeBytes, err := c.GetVolumeSize(volumeName)
			if err != nil {
				size = "Unknown"
				sizeBytes = 0
			}

			volumeMap[volumeName] = &VolumeInfo{
				Name:       volumeName,
				Container:  containerName,
				MountPath:  mountPath,
				Size:       size,
				SizeBytes:  sizeBytes,
				Selected:   true, // Default to selected
			}
		}
	}

	// Convert map to slice
	var result []VolumeInfo
	for _, info := range volumeMap {
		result = append(result, *info)
	}

	return result, nil
}

// parseSizeToBytes converts size string (e.g., "1.2GB", "500MB") to bytes
func parseSizeToBytes(sizeStr string) int64 {
	// Remove any whitespace
	sizeStr = strings.TrimSpace(sizeStr)

	// Use package-level compiled regex for performance
	matches := sizeRegex.FindStringSubmatch(sizeStr)

	if len(matches) < 3 {
		return 0
	}

	var value float64
	fmt.Sscanf(matches[1], "%f", &value)

	unit := strings.ToUpper(matches[2])

	multiplier := int64(1)
	switch {
	case strings.HasPrefix(unit, "K"):
		multiplier = 1024
	case strings.HasPrefix(unit, "M"):
		multiplier = 1024 * 1024
	case strings.HasPrefix(unit, "G"):
		multiplier = 1024 * 1024 * 1024
	case strings.HasPrefix(unit, "T"):
		multiplier = 1024 * 1024 * 1024 * 1024
	}

	return int64(value * float64(multiplier))
}
