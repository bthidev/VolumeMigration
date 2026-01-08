package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/spf13/cobra"
	"volume-migrator/internal/migrator"
)

// Version information (injected at build time via ldflags)
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var (
	// CLI flags
	remoteHost            string
	interactive           bool
	sshKeyPath            string
	sshPort               string
	tempDir               string
	remoteTempDir         string
	verbose               bool
	dryRun                bool
	noCleanup             bool
	showProgress          bool
	strictHostKeyChecking bool
	acceptHostKey         bool
	knownHostsFile        string
	validateOnly          bool
	force                 bool
)

var rootCmd = &cobra.Command{
	Use:   "volume-migrator [container1] [container2...]",
	Short: "Migrate Docker volumes from local containers to a remote machine",
	Long: `Volume Migrator is a CLI tool for migrating Docker volumes from local containers to a remote Linux machine.

It automatically detects sudo requirements, supports interactive volume selection, and provides progress tracking during transfer.`,
	Example: `  # Migrate all volumes from a container
  volume-migrator mycontainer --remote user@192.168.1.100

  # Interactive mode - select which volumes to migrate
  volume-migrator mycontainer --remote user@host --interactive

  # Multiple containers with custom SSH key
  volume-migrator web-app db-server --remote user@host --ssh-key ~/.ssh/deploy_key

  # Verbose mode with dry-run
  volume-migrator app --remote user@host --verbose --dry-run`,
	Args: cobra.MinimumNArgs(1),
	RunE: runMigration,
}

func init() {
	// Required flags
	rootCmd.Flags().StringVarP(&remoteHost, "remote", "r", "", "Remote host in format user@host[:port] (required)")
	rootCmd.MarkFlagRequired("remote")

	// Optional flags
	rootCmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Display volumes and let user select which to migrate")
	rootCmd.Flags().StringVar(&sshKeyPath, "ssh-key", "", "Path to SSH private key (default: auto-detect)")
	rootCmd.Flags().StringVar(&sshPort, "ssh-port", "22", "SSH port")
	rootCmd.Flags().StringVar(&tempDir, "temp-dir", "", "Local temporary directory (default: /tmp/volume-migration-{timestamp})")
	rootCmd.Flags().StringVar(&remoteTempDir, "remote-temp-dir", "", "Remote temporary directory (default: /tmp/volume-migration-{timestamp})")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	rootCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be done without doing it")
	rootCmd.Flags().BoolVar(&validateOnly, "validate-only", false, "Validate configuration without running migration")
	rootCmd.Flags().BoolVar(&force, "force", false, "Skip disk space validation checks")
	rootCmd.Flags().BoolVar(&noCleanup, "no-cleanup", false, "Keep temporary files for debugging")
	rootCmd.Flags().BoolVarP(&showProgress, "progress", "p", true, "Show progress bars during transfer")

	// SSH security flags
	rootCmd.Flags().BoolVar(&strictHostKeyChecking, "strict-host-key-checking", true, "Verify SSH host keys against known_hosts")
	rootCmd.Flags().BoolVar(&acceptHostKey, "accept-host-key", false, "Automatically accept and add unknown host keys (DANGEROUS - use only in trusted environments)")
	rootCmd.Flags().StringVar(&knownHostsFile, "known-hosts-file", "", "Path to known_hosts file (default: ~/.ssh/known_hosts)")
}

func runMigration(cmd *cobra.Command, args []string) error {
	// Create context with cancellation support (Ctrl+C)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\n\nReceived interrupt signal. Cleaning up...")
		cancel()
	}()

	// Create migration config
	config := &migrator.Config{
		Containers:            args,
		RemoteHost:            remoteHost,
		SSHKeyPath:            sshKeyPath,
		SSHPort:               sshPort,
		TempDir:               tempDir,
		RemoteTempDir:         remoteTempDir,
		Interactive:           interactive,
		Verbose:               verbose,
		DryRun:                dryRun,
		NoCleanup:             noCleanup,
		ShowProgress:          showProgress,
		StrictHostKeyChecking: strictHostKeyChecking,
		AcceptHostKey:         acceptHostKey,
		KnownHostsFile:        knownHostsFile,
		Force:                 force,
	}

	// Validate configuration
	if err := migrator.ValidateConfig(config); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// If validate-only mode, exit after successful validation
	if validateOnly {
		fmt.Println("✓ Configuration is valid")
		fmt.Printf("  Containers: %v\n", config.Containers)
		fmt.Printf("  Remote Host: %s\n", config.RemoteHost)
		fmt.Printf("  SSH Port: %s\n", config.SSHPort)
		if config.SSHKeyPath != "" {
			fmt.Printf("  SSH Key: %s\n", config.SSHKeyPath)
		}
		if config.TempDir != "" {
			fmt.Printf("  Temp Directory: %s\n", config.TempDir)
		}
		if config.RemoteTempDir != "" {
			fmt.Printf("  Remote Temp Directory: %s\n", config.RemoteTempDir)
		}
		fmt.Printf("  Strict Host Key Checking: %v\n", config.StrictHostKeyChecking)
		return nil
	}

	// Create migrator
	m, err := migrator.NewMigrator(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}

	// Run migration
	if err := m.Migrate(); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	return nil
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  "Display version, build commit, and build date information for volume-migrator",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("volume-migrator version %s\n", version)
		fmt.Printf("  Commit:     %s\n", commit)
		fmt.Printf("  Built:      %s\n", date)
		fmt.Printf("  Go version: %s\n", runtime.Version())
		fmt.Printf("  OS/Arch:    %s/%s\n", runtime.GOOS, runtime.GOARCH)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
