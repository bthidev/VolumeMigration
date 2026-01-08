package ssh

import (
	"fmt"
	"net"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// getAuthMethods returns SSH authentication methods in priority order:
// 1. SSH Agent (if available)
// 2. Private keys from ~/.ssh/
// 3. Custom key path (if provided)
func getAuthMethods(customKeyPath string) ([]ssh.AuthMethod, error) {
	var methods []ssh.AuthMethod

	// 1. Try SSH Agent
	if agentMethods := trySSHAgent(); agentMethods != nil {
		methods = append(methods, agentMethods)
	}

	// 2. Try custom key path if provided
	if customKeyPath != "" {
		if key, err := loadPrivateKey(customKeyPath); err == nil {
			methods = append(methods, ssh.PublicKeys(key))
		} else {
			return nil, fmt.Errorf("failed to load custom key %s: %w", customKeyPath, err)
		}
	} else {
		// Try common private key locations
		homeDir, err := os.UserHomeDir()
		if err == nil {
			keyPaths := []string{
				"id_rsa",
				"id_ed25519",
				"id_ecdsa",
				"id_dsa",
			}

			for _, keyName := range keyPaths {
				keyPath := filepath.Join(homeDir, ".ssh", keyName)
				if key, err := loadPrivateKey(keyPath); err == nil {
					methods = append(methods, ssh.PublicKeys(key))
				}
			}
		}
	}

	if len(methods) == 0 {
		return nil, fmt.Errorf("no SSH authentication methods available")
	}

	return methods, nil
}

// trySSHAgent attempts to connect to SSH agent
func trySSHAgent() ssh.AuthMethod {
	socket := os.Getenv("SSH_AUTH_SOCK")
	if socket == "" {
		return nil
	}

	conn, err := net.Dial("unix", socket)
	if err != nil {
		return nil
	}

	agentClient := agent.NewClient(conn)
	return ssh.PublicKeysCallback(agentClient.Signers)
}

// loadPrivateKey loads a private key from a file
func loadPrivateKey(path string) (ssh.Signer, error) {
	// Validate file permissions before loading
	if err := validateKeyPermissions(path); err != nil {
		return nil, fmt.Errorf("insecure key permissions: %w", err)
	}

	key, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Try parsing without passphrase first
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		// If it's an encrypted key, we would need to handle passphrase
		// For now, we skip encrypted keys
		return nil, fmt.Errorf("key is encrypted or invalid: %w", err)
	}

	return signer, nil
}

// validateKeyPermissions checks that SSH private key has secure file permissions
// SSH keys should be readable only by owner (0400 or 0600), not by group or others
func validateKeyPermissions(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat SSH key file %s: %w", path, err)
	}

	mode := info.Mode()
	perm := mode.Perm()

	// Check if file is readable by group or others (bits 4-8)
	// SSH keys should have permissions like 0600 or 0400
	// We check if any of the group/other read/write/execute bits are set
	if perm&0077 != 0 {
		return fmt.Errorf("private key file %s has insecure permissions %o (should be 0600 or 0400)", path, perm)
	}

	return nil
}

// parseHostPort parses a host string in format "user@host:port" or "user@host"
func parseHostPort(hostStr string) (user, host, port string, err error) {
	// Default values
	port = "22"

	// Check if user is specified
	if at := findAt(hostStr); at != -1 {
		user = hostStr[:at]
		hostStr = hostStr[at+1:]
	} else {
		// Use current user if not specified
		user = os.Getenv("USER")
		if user == "" {
			user = os.Getenv("USERNAME") // Windows
		}
	}

	// Check if port is specified
	if colon := findColon(hostStr); colon != -1 {
		host = hostStr[:colon]
		port = hostStr[colon+1:]
	} else {
		host = hostStr
	}

	if user == "" {
		err = fmt.Errorf("could not determine username")
		return
	}

	if host == "" {
		err = fmt.Errorf("host cannot be empty")
		return
	}

	return
}

// findAt finds the position of @ in the string
func findAt(s string) int {
	for i, c := range s {
		if c == '@' {
			return i
		}
	}
	return -1
}

// findColon finds the position of : in the string (for port)
func findColon(s string) int {
	for i, c := range s {
		if c == ':' {
			return i
		}
	}
	return -1
}
