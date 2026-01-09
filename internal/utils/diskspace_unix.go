//go:build unix

package utils

import (
	"fmt"
	"syscall"
)

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
