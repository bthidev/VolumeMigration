package ssh

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/sftp"
	"github.com/schollz/progressbar/v3"
)

// ProgressReader wraps an io.Reader with a progress bar
type ProgressReader struct {
	io.Reader
	bar *progressbar.ProgressBar
}

// Read implements io.Reader interface with progress tracking
func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.Reader.Read(p)
	if err == nil || err == io.EOF {
		pr.bar.Add(n)
	}
	return n, err
}

// TransferFile uploads a file to the remote host via SFTP with progress tracking
func (c *Client) TransferFile(localPath, remotePath string, showProgress bool) error {
	// Open SFTP session
	sftpClient, err := sftp.NewClient(c.client)
	if err != nil {
		return fmt.Errorf("failed to create SFTP client: %w", err)
	}
	defer sftpClient.Close()

	// Open local file
	srcFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %w", err)
	}
	defer srcFile.Close()

	// Get file info for size
	stat, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat local file: %w", err)
	}

	// Ensure remote directory exists
	remoteDir := filepath.Dir(remotePath)
	if err := sftpClient.MkdirAll(remoteDir); err != nil {
		return fmt.Errorf("failed to create remote directory: %w", err)
	}

	// Create remote file
	dstFile, err := sftpClient.Create(remotePath)
	if err != nil {
		return fmt.Errorf("failed to create remote file: %w", err)
	}
	defer dstFile.Close()

	// Create progress bar if requested
	var reader io.Reader = srcFile
	if showProgress {
		bar := progressbar.DefaultBytes(
			stat.Size(),
			fmt.Sprintf("Uploading %s", filepath.Base(localPath)),
		)
		reader = &ProgressReader{Reader: srcFile, bar: bar}
		defer bar.Finish()
	}

	// Copy file
	if _, err := io.Copy(dstFile, reader); err != nil {
		return fmt.Errorf("failed to transfer file: %w", err)
	}

	return nil
}

// DownloadFile downloads a file from the remote host via SFTP with progress tracking
func (c *Client) DownloadFile(remotePath, localPath string, showProgress bool) error {
	// Open SFTP session
	sftpClient, err := sftp.NewClient(c.client)
	if err != nil {
		return fmt.Errorf("failed to create SFTP client: %w", err)
	}
	defer sftpClient.Close()

	// Open remote file
	srcFile, err := sftpClient.Open(remotePath)
	if err != nil {
		return fmt.Errorf("failed to open remote file: %w", err)
	}
	defer srcFile.Close()

	// Get file info for size
	stat, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat remote file: %w", err)
	}

	// Ensure local directory exists
	localDir := filepath.Dir(localPath)
	if err := os.MkdirAll(localDir, 0755); err != nil {
		return fmt.Errorf("failed to create local directory: %w", err)
	}

	// Create local file
	dstFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer dstFile.Close()

	// Create progress bar if requested
	var reader io.Reader = srcFile
	if showProgress {
		bar := progressbar.DefaultBytes(
			stat.Size(),
			fmt.Sprintf("Downloading %s", filepath.Base(remotePath)),
		)
		reader = &ProgressReader{Reader: srcFile, bar: bar}
		defer bar.Finish()
	}

	// Copy file
	if _, err := io.Copy(dstFile, reader); err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	return nil
}

// FileExists checks if a file exists on the remote host
func (c *Client) FileExists(remotePath string) (bool, error) {
	sftpClient, err := sftp.NewClient(c.client)
	if err != nil {
		return false, fmt.Errorf("failed to create SFTP client: %w", err)
	}
	defer sftpClient.Close()

	_, err = sftpClient.Stat(remotePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// GetFileSize returns the size of a remote file
func (c *Client) GetFileSize(remotePath string) (int64, error) {
	sftpClient, err := sftp.NewClient(c.client)
	if err != nil {
		return 0, fmt.Errorf("failed to create SFTP client: %w", err)
	}
	defer sftpClient.Close()

	stat, err := sftpClient.Stat(remotePath)
	if err != nil {
		return 0, err
	}

	return stat.Size(), nil
}
