package ssh

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"golang.org/x/crypto/ssh"
	"volume-migrator/internal/shell"
)

// Client wraps SSH client operations
type Client struct {
	client     *ssh.Client
	config     *ssh.ClientConfig
	host       string
	remoteSudo bool
	ctx        context.Context
}

// ClientConfig holds SSH client configuration options
type ClientConfig struct {
	HostString            string
	CustomKeyPath         string
	StrictHostKeyChecking bool
	AcceptHostKey         bool
	KnownHostsFile        string
}

// NewClient creates a new SSH client and establishes connection
func NewClient(ctx context.Context, cfg *ClientConfig) (*Client, error) {
	hostStr := cfg.HostString
	customKeyPath := cfg.CustomKeyPath
	user, host, port, err := parseHostPort(hostStr)
	if err != nil {
		return nil, fmt.Errorf("invalid host string: %w", err)
	}

	authMethods, err := getAuthMethods(customKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get auth methods: %w", err)
	}

	// Create host key verifier
	verifier, err := NewHostKeyVerifier(cfg.StrictHostKeyChecking, cfg.AcceptHostKey, cfg.KnownHostsFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create host key verifier: %w", err)
	}

	hostKeyCallback, err := verifier.GetCallback()
	if err != nil {
		return nil, fmt.Errorf("failed to get host key callback: %w", err)
	}

	config := &ssh.ClientConfig{
		User:            user,
		Auth:            authMethods,
		HostKeyCallback: hostKeyCallback,
		Timeout:         30 * time.Second,
	}

	// Connect to remote host
	addr := fmt.Sprintf("%s:%s", host, port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", addr, err)
	}

	sshClient := &Client{
		client: client,
		config: config,
		host:   addr,
		ctx:    ctx,
	}

	// Detect if remote Docker requires sudo
	if err := sshClient.detectRemoteSudo(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to detect remote sudo: %w", err)
	}

	return sshClient, nil
}

// detectRemoteSudo detects if Docker commands on remote require sudo
func (c *Client) detectRemoteSudo() error {
	// Try without sudo
	_, err := c.RunCommand("docker ps")
	if err == nil {
		c.remoteSudo = false
		return nil
	}

	// Try with sudo
	_, err = c.RunCommand("sudo -n docker ps")
	if err != nil {
		return fmt.Errorf("docker not accessible on remote host")
	}

	c.remoteSudo = true
	return nil
}

// RunCommand executes a command on the remote host
func (c *Client) RunCommand(cmd string) (string, error) {
	session, err := c.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	if err := session.Run(cmd); err != nil {
		return "", fmt.Errorf("command failed: %w, stderr: %s", err, stderr.String())
	}

	return stdout.String(), nil
}

// RunDockerCommand executes a Docker command on the remote host
// Automatically adds sudo if required
func (c *Client) RunDockerCommand(args ...string) (string, error) {
	cmd := "docker"
	if c.remoteSudo {
		cmd = "sudo docker"
	}

	for _, arg := range args {
		cmd += " " + arg
	}

	return c.RunCommand(cmd)
}

// RunCommandWithOutput executes a command and captures stdout and stderr separately
func (c *Client) RunCommandWithOutput(cmd string, stdout, stderr *bytes.Buffer) error {
	session, err := c.client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	session.Stdout = stdout
	session.Stderr = stderr

	return session.Run(cmd)
}

// CreateDirectory creates a directory on the remote host
func (c *Client) CreateDirectory(path string) error {
	// Sanitize and escape path to prevent command injection
	safePath := shell.SanitizePathForRemote(path)
	cmd := fmt.Sprintf("mkdir -p %s", shell.ShellEscape(safePath))
	_, err := c.RunCommand(cmd)
	if err != nil {
		return fmt.Errorf("failed to create directory %s on remote host: %w", path, err)
	}
	return nil
}

// RemoveFile removes a file on the remote host
func (c *Client) RemoveFile(path string) error {
	// Sanitize and escape path to prevent command injection
	safePath := shell.SanitizePathForRemote(path)
	cmd := fmt.Sprintf("rm -f %s", shell.ShellEscape(safePath))
	_, err := c.RunCommand(cmd)
	if err != nil {
		return fmt.Errorf("failed to remove file %s on remote host: %w", path, err)
	}
	return nil
}

// RemoveDirectory removes a directory on the remote host
func (c *Client) RemoveDirectory(path string) error {
	// Sanitize and escape path to prevent command injection
	// DANGEROUS: rm -rf - extra safety checks
	safePath := shell.SanitizePathForRemote(path)

	// Extra safety: refuse to delete root or system directories
	if safePath == "/" || safePath == "/bin" || safePath == "/etc" ||
	   safePath == "/usr" || safePath == "/var" || safePath == "/home" {
		return fmt.Errorf("refusing to delete system directory: %s", safePath)
	}

	cmd := fmt.Sprintf("rm -rf %s", shell.ShellEscape(safePath))
	_, err := c.RunCommand(cmd)
	if err != nil {
		return fmt.Errorf("failed to remove directory %s on remote host: %w", path, err)
	}
	return nil
}

// RequiresSudo returns whether remote Docker commands require sudo
func (c *Client) RequiresSudo() bool {
	return c.remoteSudo
}

// Close closes the SSH connection
func (c *Client) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// GetClient returns the underlying SSH client for advanced operations
func (c *Client) GetClient() *ssh.Client {
	return c.client
}
