package errors

import (
	"fmt"
)

// VolumeNotFoundError indicates a Docker volume could not be found
type VolumeNotFoundError struct {
	VolumeName string
	Err        error
}

func (e *VolumeNotFoundError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("volume '%s' not found: %v", e.VolumeName, e.Err)
	}
	return fmt.Sprintf("volume '%s' not found", e.VolumeName)
}

func (e *VolumeNotFoundError) Unwrap() error {
	return e.Err
}

// NewVolumeNotFoundError creates a new VolumeNotFoundError.
// Use this when a Docker volume inspect or lookup fails because the volume doesn't exist.
// The err parameter can be nil if no underlying error is available.
func NewVolumeNotFoundError(volumeName string, err error) *VolumeNotFoundError {
	return &VolumeNotFoundError{
		VolumeName: volumeName,
		Err:        err,
	}
}

// SSHConnectionError indicates an SSH connection failure
type SSHConnectionError struct {
	Host string
	Err  error
}

func (e *SSHConnectionError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("failed to connect to SSH host '%s': %v", e.Host, e.Err)
	}
	return fmt.Sprintf("failed to connect to SSH host '%s'", e.Host)
}

func (e *SSHConnectionError) Unwrap() error {
	return e.Err
}

// NewSSHConnectionError creates a new SSHConnectionError.
// Use this when SSH connection establishment fails (authentication, network, or host key issues).
// The host parameter should include the user@host format if available.
// The err parameter should contain the underlying connection error.
func NewSSHConnectionError(host string, err error) *SSHConnectionError {
	return &SSHConnectionError{
		Host: host,
		Err:  err,
	}
}

// DiskSpaceError indicates insufficient disk space
type DiskSpaceError struct {
	Location  string // "local" or "remote"
	Required  int64
	Available int64
	Err       error
}

func (e *DiskSpaceError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("insufficient disk space on %s: required %d bytes, available %d bytes: %v",
			e.Location, e.Required, e.Available, e.Err)
	}
	return fmt.Sprintf("insufficient disk space on %s: required %d bytes, available %d bytes",
		e.Location, e.Required, e.Available)
}

func (e *DiskSpaceError) Unwrap() error {
	return e.Err
}

// NewDiskSpaceError creates a new DiskSpaceError.
// Use this when a volume migration would fail due to insufficient disk space.
// The location parameter should be "local" or "remote" to indicate which system
// lacks sufficient space. The required and available parameters are in bytes.
// The err parameter can be nil if the error is purely insufficient space.
func NewDiskSpaceError(location string, required, available int64, err error) *DiskSpaceError {
	return &DiskSpaceError{
		Location:  location,
		Required:  required,
		Available: available,
		Err:       err,
	}
}

// PermissionError indicates a permission-related error
type PermissionError struct {
	Operation string // e.g., "read", "write", "execute"
	Path      string
	Err       error
}

func (e *PermissionError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("permission denied: cannot %s '%s': %v", e.Operation, e.Path, e.Err)
	}
	return fmt.Sprintf("permission denied: cannot %s '%s'", e.Operation, e.Path)
}

func (e *PermissionError) Unwrap() error {
	return e.Err
}

// NewPermissionError creates a new PermissionError.
// Use this when file or directory operations fail due to insufficient permissions.
// The operation parameter should describe the action (e.g., "read", "write", "execute").
// The path parameter is the file or directory path that caused the error.
// The err parameter should contain the underlying permission error from the OS.
func NewPermissionError(operation, path string, err error) *PermissionError {
	return &PermissionError{
		Operation: operation,
		Path:      path,
		Err:       err,
	}
}
