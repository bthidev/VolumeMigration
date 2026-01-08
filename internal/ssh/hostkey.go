package ssh

import (
	"fmt"
	"net"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// HostKeyVerifier handles SSH host key verification
type HostKeyVerifier struct {
	knownHostsPath string
	strictChecking bool
	acceptNewKeys  bool
}

// NewHostKeyVerifier creates a new host key verifier with the specified security settings.
// Parameters:
//   - strictChecking: When true, rejects any unknown or changed host keys
//   - acceptNewKeys: When true, automatically accepts and saves new host keys (only with strictChecking=false)
//   - knownHostsPath: Path to known_hosts file, or empty string to use ~/.ssh/known_hosts
//
// Returns an error if the home directory cannot be determined when knownHostsPath is empty.
func NewHostKeyVerifier(strictChecking, acceptNewKeys bool, knownHostsPath string) (*HostKeyVerifier, error) {
	if knownHostsPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		knownHostsPath = filepath.Join(homeDir, ".ssh", "known_hosts")
	}

	return &HostKeyVerifier{
		knownHostsPath: knownHostsPath,
		strictChecking: strictChecking,
		acceptNewKeys:  acceptNewKeys,
	}, nil
}

// GetCallback returns the HostKeyCallback for ssh.ClientConfig.
// The callback behavior depends on the verifier configuration:
//   - When acceptNewKeys is true and strictChecking is false: Automatically
//     accepts and saves new host keys, warning on key changes (MITM detection).
//   - When strictChecking is true: Uses strict verification, rejecting any
//     unknown hosts or hosts with changed keys.
//   - Otherwise: Standard known_hosts verification behavior.
//
// Returns an error if the known_hosts file cannot be read or created.
func (v *HostKeyVerifier) GetCallback() (ssh.HostKeyCallback, error) {
	// If accept-new-keys is set and strict checking is off, use custom callback
	if v.acceptNewKeys && !v.strictChecking {
		return v.acceptNewKeyCallback()
	}

	// Use strict known_hosts verification
	callback, err := knownhosts.New(v.knownHostsPath)
	if err != nil {
		// If known_hosts doesn't exist and we're not strict, create it
		if os.IsNotExist(err) && !v.strictChecking {
			if err := v.createKnownHostsFile(); err != nil {
				return nil, fmt.Errorf("failed to create known_hosts file: %w", err)
			}
			callback, err = knownhosts.New(v.knownHostsPath)
			if err != nil {
				return nil, fmt.Errorf("failed to load known_hosts after creation: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to load known_hosts: %w", err)
		}
	}

	return callback, nil
}

// acceptNewKeyCallback creates a callback that accepts new keys and adds them
func (v *HostKeyVerifier) acceptNewKeyCallback() (ssh.HostKeyCallback, error) {
	// Ensure known_hosts exists
	if _, err := os.Stat(v.knownHostsPath); os.IsNotExist(err) {
		if err := v.createKnownHostsFile(); err != nil {
			return nil, fmt.Errorf("failed to create known_hosts: %w", err)
		}
	}

	callback, err := knownhosts.New(v.knownHostsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load known_hosts: %w", err)
	}

	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		err := callback(hostname, remote, key)
		if err != nil {
			// Check if it's a key mismatch (security issue) or unknown host
			keyErr, isKeyErr := err.(*knownhosts.KeyError)

			if isKeyErr && len(keyErr.Want) > 0 {
				// Host key has changed - potential MITM attack
				return fmt.Errorf("WARNING: REMOTE HOST IDENTIFICATION HAS CHANGED!\n"+
					"IT IS POSSIBLE THAT SOMEONE IS DOING SOMETHING NASTY!\n"+
					"Host key for %s has changed.\n"+
					"Remove old key from %s and try again.\n"+
					"Or use ssh-keygen -R %s\n%w",
					hostname, v.knownHostsPath, hostname, err)
			}

			// Unknown host - add it if acceptNewKeys is true
			fmt.Fprintf(os.Stderr, "WARNING: Unknown host %s\n", hostname)
			fmt.Fprintf(os.Stderr, "Fingerprint: %s\n", ssh.FingerprintSHA256(key))
			fmt.Fprintf(os.Stderr, "Adding new host key to %s\n", v.knownHostsPath)

			if err := v.addHostKey(hostname, key); err != nil {
				return fmt.Errorf("failed to add host key: %w", err)
			}
			return nil
		}
		// Unexpected error from host key verification
		return fmt.Errorf("unexpected host key verification error for %s: %w", hostname, err)
	}, nil
}

// createKnownHostsFile creates an empty known_hosts file with proper permissions
func (v *HostKeyVerifier) createKnownHostsFile() error {
	dir := filepath.Dir(v.knownHostsPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create .ssh directory: %w", err)
	}

	file, err := os.OpenFile(v.knownHostsPath, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to create known_hosts file: %w", err)
	}
	return file.Close()
}

// addHostKey adds a host key to known_hosts
func (v *HostKeyVerifier) addHostKey(hostname string, key ssh.PublicKey) error {
	file, err := os.OpenFile(v.knownHostsPath, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open known_hosts: %w", err)
	}
	defer file.Close()

	// Format: hostname keytype base64key
	line := knownhosts.Line([]string{hostname}, key)
	if _, err := file.WriteString(line + "\n"); err != nil {
		return fmt.Errorf("failed to write to known_hosts: %w", err)
	}

	return nil
}

// GetFingerprint returns the SHA256 fingerprint of a public key.
// The fingerprint is formatted as "SHA256:base64hash" and can be used to
// verify host identity when accepting new keys or diagnosing MITM attacks.
// This is the same format displayed by OpenSSH commands like ssh-keygen.
func GetFingerprint(key ssh.PublicKey) string {
	return ssh.FingerprintSHA256(key)
}
