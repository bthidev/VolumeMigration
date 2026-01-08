package migrator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"volume-migrator/internal/docker"
	"volume-migrator/internal/ssh"
	"volume-migrator/internal/ui"
	"volume-migrator/internal/utils"
)

var log *logrus.Logger

func init() {
	log = utils.GetLogger()
}

// Config holds migration configuration
type Config struct {
	Containers            []string
	RemoteHost            string
	SSHKeyPath            string
	SSHPort               string
	TempDir               string
	RemoteTempDir         string
	Interactive           bool
	Verbose               bool
	DryRun                bool
	NoCleanup             bool
	ShowProgress          bool
	StrictHostKeyChecking bool
	AcceptHostKey         bool
	KnownHostsFile        string
	Force                 bool
}

// ValidateConfig validates the migration configuration
func ValidateConfig(config *Config) error {
	// Validate containers are non-empty
	if len(config.Containers) == 0 {
		return fmt.Errorf("no containers specified")
	}

	// Validate each container name is non-empty
	for i, container := range config.Containers {
		if strings.TrimSpace(container) == "" {
			return fmt.Errorf("container at index %d is empty", i)
		}
	}

	// Validate remote host format (user@host or user@host:port)
	if config.RemoteHost == "" {
		return fmt.Errorf("remote host not specified")
	}

	// Check for @ symbol (user@host format)
	if !strings.Contains(config.RemoteHost, "@") {
		return fmt.Errorf("remote host must be in format 'user@host' or 'user@host:port', got: %s", config.RemoteHost)
	}

	// Extract user and host parts
	parts := strings.Split(config.RemoteHost, "@")
	if len(parts) != 2 {
		return fmt.Errorf("invalid remote host format: %s", config.RemoteHost)
	}

	user := strings.TrimSpace(parts[0])
	hostPart := strings.TrimSpace(parts[1])

	if user == "" {
		return fmt.Errorf("username cannot be empty in remote host: %s", config.RemoteHost)
	}

	if hostPart == "" {
		return fmt.Errorf("host cannot be empty in remote host: %s", config.RemoteHost)
	}

	// Validate SSH port if specified
	if config.SSHPort != "" {
		port, err := strconv.Atoi(config.SSHPort)
		if err != nil {
			return fmt.Errorf("invalid SSH port '%s': must be a number", config.SSHPort)
		}
		if port < 1 || port > 65535 {
			return fmt.Errorf("invalid SSH port %d: must be between 1 and 65535", port)
		}
	}

	// Validate temp directories are absolute paths (if specified)
	if config.TempDir != "" && !filepath.IsAbs(config.TempDir) {
		return fmt.Errorf("temp directory must be an absolute path: %s", config.TempDir)
	}

	if config.RemoteTempDir != "" && !filepath.IsAbs(config.RemoteTempDir) {
		return fmt.Errorf("remote temp directory must be an absolute path: %s", config.RemoteTempDir)
	}

	// Validate conflicting flags
	if config.StrictHostKeyChecking && config.AcceptHostKey {
		return fmt.Errorf("conflicting flags: --strict-host-key-checking and --accept-host-key cannot both be enabled")
	}

	// Validate SSH key path exists if specified
	if config.SSHKeyPath != "" {
		if _, err := os.Stat(config.SSHKeyPath); os.IsNotExist(err) {
			return fmt.Errorf("SSH key file does not exist: %s", config.SSHKeyPath)
		}
	}

	// Validate known_hosts file exists if specified and strict checking is enabled
	if config.KnownHostsFile != "" && config.StrictHostKeyChecking {
		if _, err := os.Stat(config.KnownHostsFile); os.IsNotExist(err) {
			return fmt.Errorf("known_hosts file does not exist: %s (use --accept-host-key to create it)", config.KnownHostsFile)
		}
	}

	return nil
}

// Migrator orchestrates the volume migration process
type Migrator struct {
	config       *Config
	dockerClient *docker.Client
	sshClient    *ssh.Client
	ctx          context.Context
}

// NewMigrator creates a new migrator instance
func NewMigrator(ctx context.Context, config *Config) (*Migrator, error) {
	if len(config.Containers) == 0 {
		return nil, fmt.Errorf("no containers specified")
	}

	if config.RemoteHost == "" {
		return nil, fmt.Errorf("remote host not specified")
	}

	// Set default temp directories if not specified
	if config.TempDir == "" {
		config.TempDir = filepath.Join(os.TempDir(), fmt.Sprintf("volume-migration-%d", time.Now().Unix()))
	}

	if config.RemoteTempDir == "" {
		config.RemoteTempDir = fmt.Sprintf("/tmp/volume-migration-%d", time.Now().Unix())
	}

	return &Migrator{
		config: config,
		ctx:    ctx,
	}, nil
}

// Migrate performs the complete migration process
func (m *Migrator) Migrate() error {
	// Set verbose logging
	utils.SetVerbose(m.config.Verbose)

	// Phase 1: Initialize Docker client
	log.Info("=== Phase 1: Initialization ===")

	dockerClient, err := docker.NewClient(m.ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize Docker client: %w", err)
	}
	m.dockerClient = dockerClient

	log.WithField("requires_sudo", dockerClient.RequiresSudo()).Debug("Local Docker sudo detection complete")

	// Phase 2: Establish SSH connection
	log.WithField("remote_host", m.config.RemoteHost).Info("Connecting to remote host")

	sshConfig := &ssh.ClientConfig{
		HostString:            m.config.RemoteHost,
		CustomKeyPath:         m.config.SSHKeyPath,
		StrictHostKeyChecking: m.config.StrictHostKeyChecking,
		AcceptHostKey:         m.config.AcceptHostKey,
		KnownHostsFile:        m.config.KnownHostsFile,
	}

	sshClient, err := ssh.NewClient(m.ctx, sshConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to remote host: %w", err)
	}
	m.sshClient = sshClient
	defer sshClient.Close()

	log.WithField("requires_sudo", sshClient.RequiresSudo()).Debug("Remote Docker sudo detection complete")

	// Phase 3: Discover volumes
	log.Info("=== Phase 2: Volume Discovery ===")

	volumes, err := m.discoverVolumes()
	if err != nil {
		return fmt.Errorf("failed to discover volumes: %w", err)
	}

	if len(volumes) == 0 {
		log.Warn("No volumes found to migrate")
		return nil
	}

	// Phase 4: Interactive selection (if enabled)
	if m.config.Interactive {
		log.Info("=== Phase 2.5: Volume Selection ===")

		selectedVolumes, err := ui.SelectVolumes(volumes)
		if err != nil {
			return fmt.Errorf("volume selection failed: %w", err)
		}
		volumes = selectedVolumes
	} else {
		// Display volumes that will be migrated
		ui.DisplayVolumeTable(volumes)
	}

	// Phase 4.5: Disk space validation
	if !m.config.Force {
		log.Debug("Validating disk space requirements")

		// Calculate total required space
		var totalVolumeSize int64
		for _, v := range volumes {
			totalVolumeSize += v.SizeBytes
		}

		estimatedArchiveSize := utils.CalculateRequiredSpace(totalVolumeSize)
		log.WithFields(logrus.Fields{
			"total_volume_size": utils.FormatBytes(totalVolumeSize),
			"estimated_archive": utils.FormatBytes(estimatedArchiveSize),
		}).Debug("Calculated space requirements")

		// Check local disk space
		localSpace, err := utils.GetLocalDiskSpace(m.config.TempDir)
		if err != nil {
			if m.config.Verbose {
				log.WithError(err).Warn("Could not check local disk space")
			}
		} else {
			log.WithFields(logrus.Fields{
				"available": utils.FormatBytes(int64(localSpace.Available)),
				"required":  utils.FormatBytes(estimatedArchiveSize),
			}).Debug("Local disk space check")

			if err := utils.ValidateDiskSpace("local", uint64(estimatedArchiveSize), localSpace.Available); err != nil {
				return fmt.Errorf("%w (use --force to override)", err)
			}
		}

		// Check remote disk space
		remoteSpace, err := utils.GetRemoteDiskSpace(m.sshClient, m.config.RemoteTempDir)
		if err != nil {
			if m.config.Verbose {
				log.WithError(err).Warn("Could not check remote disk space")
			}
		} else {
			log.WithFields(logrus.Fields{
				"available": utils.FormatBytes(int64(remoteSpace.Available)),
				"required":  utils.FormatBytes(estimatedArchiveSize),
			}).Debug("Remote disk space check")

			if err := utils.ValidateDiskSpace("remote", uint64(estimatedArchiveSize), remoteSpace.Available); err != nil {
				return fmt.Errorf("%w (use --force to override)", err)
			}
		}

		log.Debug("Disk space validation passed")
	} else {
		log.Warn("Skipping disk space validation (--force enabled)")
	}

	if m.config.DryRun {
		log.WithField("volume_count", len(volumes)).Info("Dry run mode: No actual migration will be performed")
		return nil
	}

	// Extract volume names
	volumeNames := make([]string, len(volumes))
	for i, v := range volumes {
		volumeNames[i] = v.Name
	}

	// Phase 5: Export volumes
	log.Info("=== Phase 3: Export Volumes ===")

	archivePaths, err := m.exportVolumes(volumeNames)
	if err != nil {
		return fmt.Errorf("failed to export volumes: %w", err)
	}

	// Setup cleanup on exit if not disabled
	if !m.config.NoCleanup {
		defer func() {
			log.Debug("=== Phase 6: Cleanup ===")
			if err := CleanupLocal(m.config.TempDir); err != nil {
				log.WithError(err).Error("Failed to cleanup local temporary directory")
			}
			if err := CleanupRemote(m.sshClient, m.config.RemoteTempDir); err != nil {
				log.WithError(err).Error("Failed to cleanup remote temporary directory")
			}
		}()
	}

	// Phase 6: Transfer volumes
	log.Debug("=== Phase 4: Transfer Archives ===")

	if err := m.transferVolumes(archivePaths); err != nil {
		return fmt.Errorf("failed to transfer volumes: %w", err)
	}

	// Phase 7: Import volumes on remote
	log.Debug("=== Phase 5: Import Volumes ===")

	if err := m.importVolumes(archivePaths); err != nil {
		return fmt.Errorf("failed to import volumes: %w", err)
	}

	log.WithFields(logrus.Fields{
		"volumes":     len(volumeNames),
		"remote_host": m.config.RemoteHost,
	}).Info("Migration completed successfully")

	return nil
}

// discoverVolumes discovers all volumes from specified containers
func (m *Migrator) discoverVolumes() ([]docker.VolumeInfo, error) {
	volumes, err := m.dockerClient.GetAllVolumesInfo(m.config.Containers)
	if err != nil {
		return nil, err
	}

	log.WithFields(logrus.Fields{
		"volumes":    len(volumes),
		"containers": len(m.config.Containers),
	}).Debug("Volume discovery complete")

	return volumes, nil
}

// exportVolumes exports all volumes to local archives
func (m *Migrator) exportVolumes(volumeNames []string) (map[string]string, error) {
	// Create temp directory
	if err := os.MkdirAll(m.config.TempDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	return ExportVolumes(m.dockerClient, volumeNames, m.config.TempDir)
}

// transferVolumes transfers archive files to remote host
func (m *Migrator) transferVolumes(archivePaths map[string]string) error {
	// Create remote temp directory
	if err := m.sshClient.CreateDirectory(m.config.RemoteTempDir); err != nil {
		return fmt.Errorf("failed to create remote temp directory: %w", err)
	}

	// Transfer each archive
	for volumeName, localPath := range archivePaths {
		remotePath := filepath.Join(m.config.RemoteTempDir, filepath.Base(localPath))

		log.WithField("volume", volumeName).Debug("Transferring volume")

		if err := m.sshClient.TransferFile(localPath, remotePath, m.config.ShowProgress); err != nil {
			return fmt.Errorf("failed to transfer volume %s: %w", volumeName, err)
		}
	}

	return nil
}

// importVolumes imports volumes on remote host
func (m *Migrator) importVolumes(archivePaths map[string]string) error {
	return ImportVolumes(m.sshClient, archivePaths, m.config.RemoteTempDir)
}
