package migrator

import (
	"fmt"
	"path/filepath"

	"volume-migrator/internal/shell"
	"volume-migrator/internal/ssh"
)

// ImportVolume imports a volume archive on the remote machine
// Creates a Docker volume and populates it with data from the archive
func ImportVolume(sshClient *ssh.Client, volumeName, archivePath string) error {
	// Validate volume name to prevent command injection
	if !shell.ValidateVolumeName(volumeName) {
		return fmt.Errorf("invalid volume name '%s': must contain only alphanumeric characters, dashes, underscores, and dots", volumeName)
	}

	log.WithField("volume", volumeName).Debug("Importing volume on remote host")

	// Step 1: Create the volume on remote
	createCmd := fmt.Sprintf("volume create %s", volumeName)
	if _, err := sshClient.RunDockerCommand(createCmd); err != nil {
		return fmt.Errorf("failed to create volume %s on remote: %w", volumeName, err)
	}

	log.WithField("volume", volumeName).Debug("Created volume on remote")

	// Step 2: Extract archive data into the volume
	// Get the directory and filename from archive path
	archiveDir := filepath.Dir(archivePath)
	archiveFile := filepath.Base(archivePath)

	// Build docker command to import
	// Note: On remote, we need to escape the command properly
	importCmd := fmt.Sprintf(
		`run --rm -v %s:/data -v %s:/backup alpine tar xzf /backup/%s -C /data`,
		volumeName, archiveDir, archiveFile,
	)

	if _, err := sshClient.RunDockerCommand(importCmd); err != nil {
		// Cleanup: remove the volume we just created
		if _, cleanupErr := sshClient.RunDockerCommand(fmt.Sprintf("volume rm %s", volumeName)); cleanupErr != nil {
			log.WithField("volume", volumeName).WithError(cleanupErr).Warn("Failed to cleanup volume after import failure")
		}
		return fmt.Errorf("failed to import data into volume %s: %w", volumeName, err)
	}

	log.WithField("volume", volumeName).Debug("Successfully imported volume")

	return nil
}

// ImportVolumes imports multiple volumes from archives on the remote machine
func ImportVolumes(sshClient *ssh.Client, archivePaths map[string]string, remoteTempDir string) error {
	for volumeName, archivePath := range archivePaths {
		// Construct remote archive path
		remoteArchivePath := filepath.Join(remoteTempDir, filepath.Base(archivePath))

		if err := ImportVolume(sshClient, volumeName, remoteArchivePath); err != nil {
			return fmt.Errorf("failed to import volume %s: %w", volumeName, err)
		}
	}

	return nil
}

// VerifyVolumeExists checks if a volume exists on the remote host
func VerifyVolumeExists(sshClient *ssh.Client, volumeName string) (bool, error) {
	output, err := sshClient.RunDockerCommand(fmt.Sprintf("volume inspect %s", volumeName))
	if err != nil {
		return false, nil
	}

	return len(output) > 0, nil
}
