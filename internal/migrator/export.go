package migrator

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"volume-migrator/internal/docker"
	"volume-migrator/internal/shell"
	"volume-migrator/internal/utils"
)

// ExportVolume exports a Docker volume to a tar.gz archive
// Uses a temporary Alpine container to access and compress the volume data
func ExportVolume(dockerClient *docker.Client, volumeName, outputPath string) error {
	// Validate volume name to prevent command injection and path traversal
	if !shell.ValidateVolumeName(volumeName) {
		return fmt.Errorf("invalid volume name '%s': must contain only alphanumeric characters, dashes, underscores, and dots", volumeName)
	}

	log.WithFields(logrus.Fields{
		"volume":      volumeName,
		"output_path": outputPath,
	}).Debug("Exporting volume")

	// Ensure output directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Construct docker command to export volume
	// Mount volume as read-only to avoid conflicts with running containers
	args := []string{
		"run", "--rm",
		"-v", fmt.Sprintf("%s:/data:ro", volumeName),
		"-v", fmt.Sprintf("%s:/backup", outputDir),
		"alpine",
		"tar", "czf", fmt.Sprintf("/backup/%s", filepath.Base(outputPath)),
		"-C", "/data", ".",
	}

	var stdout, stderr bytes.Buffer
	if err := dockerClient.ExecCommandWithOutput(&stdout, &stderr, args...); err != nil {
		return fmt.Errorf("failed to export volume %s: %w, stderr: %s", volumeName, err, stderr.String())
	}

	// Verify archive was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		return fmt.Errorf("archive %s was not created", outputPath)
	}

	// Get archive size
	stat, _ := os.Stat(outputPath)
	log.WithFields(logrus.Fields{
		"volume": volumeName,
		"size":   utils.FormatBytes(stat.Size()),
	}).Debug("Successfully exported volume")

	return nil
}

// ExportVolumes exports multiple volumes to a directory
func ExportVolumes(dockerClient *docker.Client, volumes []string, outputDir string) (map[string]string, error) {
	archivePaths := make(map[string]string)

	for _, volumeName := range volumes {
		archivePath := filepath.Join(outputDir, fmt.Sprintf("%s.tar.gz", volumeName))

		if err := ExportVolume(dockerClient, volumeName, archivePath); err != nil {
			return nil, fmt.Errorf("failed to export volume %s: %w", volumeName, err)
		}

		archivePaths[volumeName] = archivePath
	}

	return archivePaths, nil
}
