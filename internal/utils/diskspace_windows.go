//go:build windows

package utils

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	kernel32         = syscall.NewLazyDLL("kernel32.dll")
	getDiskFreeSpace = kernel32.NewProc("GetDiskFreeSpaceExW")
)

// GetLocalDiskSpace returns disk space information for a local path on Windows
func GetLocalDiskSpace(path string) (*DiskSpaceInfo, error) {
	var freeBytesAvailable, totalBytes, totalFreeBytes uint64

	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return nil, fmt.Errorf("failed to convert path to UTF16: %w", err)
	}

	ret, _, err := getDiskFreeSpace.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		uintptr(unsafe.Pointer(&freeBytesAvailable)),
		uintptr(unsafe.Pointer(&totalBytes)),
		uintptr(unsafe.Pointer(&totalFreeBytes)),
	)

	if ret == 0 {
		return nil, fmt.Errorf("failed to get disk space for %s: %w", path, err)
	}

	used := totalBytes - totalFreeBytes

	return &DiskSpaceInfo{
		Total:     totalBytes,
		Available: freeBytesAvailable,
		Used:      used,
	}, nil
}
