package utils

import (
	"fmt"
	"strconv"
	"strings"
	"syscall"

	"volume-migrator/internal/ssh"
)

// DiskSpaceInfo holds disk space information
type DiskSpaceInfo struct {
	Total     uint64
	Available uint64
	Used      uint64
}

// GetLocalDiskSpace returns disk space information for a local path
func GetLocalDiskSpace(path string) (*DiskSpaceInfo, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return nil, fmt.Errorf("failed to get disk space for %s: %w", path, err)
	}

	// Calculate space in bytes
	total := stat.Blocks * uint64(stat.Bsize)
	available := stat.Bavail * uint64(stat.Bsize)
	used := total - (stat.Bfree * uint64(stat.Bsize))

	return &DiskSpaceInfo{
		Total:     total,
		Available: available,
		Used:      used,
	}, nil
}

// GetRemoteDiskSpace returns disk space information for a remote path via SSH
func GetRemoteDiskSpace(sshClient *ssh.Client, remotePath string) (*DiskSpaceInfo, error) {
	// Use df -k to get disk space in kilobytes
	// -P flag ensures POSIX output format (single line per filesystem)
	cmd := fmt.Sprintf("df -Pk %s", remotePath)
	output, err := sshClient.RunCommand(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get remote disk space: %w", err)
	}

	// Parse df output
	// Expected format:
	// Filesystem     1024-blocks      Used Available Capacity Mounted on
	// /dev/sda1        10000000   5000000   4500000      53% /
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("unexpected df output: %s", output)
	}

	// Parse the second line (data line)
	fields := strings.Fields(lines[1])
	if len(fields) < 4 {
		return nil, fmt.Errorf("unexpected df output format: %s", lines[1])
	}

	// Parse values (in KB)
	totalKB, err := strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse total size: %w", err)
	}

	usedKB, err := strconv.ParseUint(fields[2], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse used size: %w", err)
	}

	availableKB, err := strconv.ParseUint(fields[3], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse available size: %w", err)
	}

	// Convert KB to bytes
	return &DiskSpaceInfo{
		Total:     totalKB * 1024,
		Available: availableKB * 1024,
		Used:      usedKB * 1024,
	}, nil
}

// CalculateRequiredSpace estimates required space for volume export
// Uses conservative estimate assuming minimal compression for safety
func CalculateRequiredSpace(volumeSizeBytes int64) int64 {
	// Conservative estimate: assume no compression (1.0x ratio)
	// This ensures we have sufficient space even if data doesn't compress well
	// (e.g., already compressed files, encrypted data, random data)

	// Add 10% buffer for:
	// - Filesystem overhead and metadata
	// - Temporary files during compression
	// - Safety margin
	buffer := 1.10
	return int64(float64(volumeSizeBytes) * buffer)
}

// ValidateDiskSpace checks if there's sufficient disk space for the operation
func ValidateDiskSpace(location string, required, available uint64) error {
	if available < required {
		return fmt.Errorf("insufficient disk space on %s: required %d bytes (%s), available %d bytes (%s)",
			location,
			required, FormatBytes(int64(required)),
			available, FormatBytes(int64(available)))
	}
	return nil
}

// FormatBytes formats bytes into human-readable string
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
