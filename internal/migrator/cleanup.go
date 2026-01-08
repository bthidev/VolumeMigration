package migrator

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"volume-migrator/internal/ssh"
)

// CleanupLocal removes local temporary files and directories
func CleanupLocal(tempDir string) error {
	log.WithField("temp_dir", tempDir).Debug("Cleaning up local temporary directory")

	if err := os.RemoveAll(tempDir); err != nil {
		return fmt.Errorf("failed to clean up local temp directory: %w", err)
	}

	return nil
}

// CleanupRemote removes remote temporary files and directories
func CleanupRemote(sshClient *ssh.Client, remoteTempDir string) error {
	log.WithField("remote_temp_dir", remoteTempDir).Debug("Cleaning up remote temporary directory")

	if err := sshClient.RemoveDirectory(remoteTempDir); err != nil {
		return fmt.Errorf("failed to clean up remote temp directory: %w", err)
	}

	return nil
}

// CleanupArchives removes specific archive files
func CleanupArchives(archivePaths map[string]string) error {
	for volumeName, path := range archivePaths {
		log.WithFields(logrus.Fields{
			"volume": volumeName,
			"path":   path,
		}).Debug("Removing archive")

		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove archive for volume %s: %w", volumeName, err)
		}
	}

	return nil
}

// CleanupRemoteArchives removes specific archive files on remote
func CleanupRemoteArchives(sshClient *ssh.Client, archivePaths map[string]string, remoteTempDir string) error {
	for volumeName, path := range archivePaths {
		remotePath := fmt.Sprintf("%s/%s", remoteTempDir, path)

		log.WithFields(logrus.Fields{
			"volume":      volumeName,
			"remote_path": remotePath,
		}).Debug("Removing remote archive")

		if err := sshClient.RemoveFile(remotePath); err != nil {
			return fmt.Errorf("failed to remove remote archive for volume %s: %w", volumeName, err)
		}
	}

	return nil
}
