# Docker Volume Migrator

[![CI](https://github.com/yourusername/volume-migrator/workflows/CI/badge.svg)](https://github.com/yourusername/volume-migrator/actions)
[![codecov](https://codecov.io/gh/yourusername/volume-migrator/branch/main/graph/badge.svg)](https://codecov.io/gh/yourusername/volume-migrator)
[![Go Report Card](https://goreportcard.com/badge/github.com/yourusername/volume-migrator)](https://goreportcard.com/report/github.com/yourusername/volume-migrator)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21%2B-blue)](https://go.dev/)

A secure, production-ready CLI tool for migrating Docker volumes from local containers to remote Linux machines. Built with Go for easy deployment as a single binary.

## Features

- **Container-based volume discovery**: Specify one or more container names to discover their volumes
- **Interactive selection mode**: Display all discovered volumes with details (size, mount path, container) and let users choose which to migrate
- **Automatic mode**: Migrate all discovered volumes without prompting (default)
- **Sudo auto-detection**: Automatically detects if sudo is required for Docker commands on both local and remote systems
- **Progress tracking**: Real-time progress bars for volume export and transfer operations
- **SSH host key verification**: Secure SSH connections with known_hosts verification (MITM attack prevention)
- **Configuration validation**: Validate configuration before running migration with `--validate-only`
- **Disk space validation**: Checks available disk space before migration to prevent failures
- **Structured logging**: Comprehensive logging with logrus for better observability
- **Single binary deployment**: Cross-compiled Go binary for easy distribution
- **SSH/SFTP transfer**: Secure file transfer using SSH keys or password authentication
- **Graceful cleanup**: Removes temporary files automatically (can be disabled for debugging)

## Installation

### Download Pre-built Binary

Download the latest release for your platform from the [Releases](https://github.com/yourusername/volume-migrator/releases) page.

```bash
# Linux AMD64
wget https://github.com/yourusername/volume-migrator/releases/latest/download/volume-migrator-linux-amd64.tar.gz
tar xzf volume-migrator-linux-amd64.tar.gz
chmod +x volume-migrator-linux-amd64
sudo mv volume-migrator-linux-amd64 /usr/local/bin/volume-migrator

# macOS ARM64 (Apple Silicon)
wget https://github.com/yourusername/volume-migrator/releases/latest/download/volume-migrator-darwin-arm64.tar.gz
tar xzf volume-migrator-darwin-arm64.tar.gz
chmod +x volume-migrator-darwin-arm64
sudo mv volume-migrator-darwin-arm64 /usr/local/bin/volume-migrator
```

### From Source

```bash
# Clone the repository
git clone <repository-url>
cd VolumeMigration

# Build for Linux
make build-linux

# Build for all platforms
make build-all

# Build for current platform
make build
```

### Using Docker

```bash
# Build the Docker image
docker build -t volume-migrator:latest .

# Run with Docker
docker run --rm \
  -v /var/run/docker.sock:/var/run/docker.sock:ro \
  -v ~/.ssh:/home/migrator/.ssh:ro \
  volume-migrator:latest mycontainer --remote user@host --dry-run

# Using docker-compose
docker-compose run --rm volume-migrator mycontainer --remote user@host --verbose
```

## Prerequisites

- Docker installed on both local and remote machines
- SSH access to the remote machine
- SSH key-based authentication configured (or password authentication as fallback)
- Alpine Docker image available (used for volume export/import)
- Sufficient disk space on both local and remote machines

## Usage

### Quick Start

```bash
# Basic migration
volume-migrator mycontainer --remote user@192.168.1.100

# Check version
volume-migrator version

# Validate configuration
volume-migrator mycontainer --remote user@host --validate-only

# Interactive mode with verbose output
volume-migrator mycontainer --remote user@host --interactive --verbose
```

### Basic Usage

Migrate all volumes from a container to a remote machine:

```bash
volume-migrator mycontainer --remote user@192.168.1.100
```

### Interactive Mode

Display volumes and select which ones to migrate:

```bash
volume-migrator mycontainer --remote user@host --interactive
```

Interactive mode shows:
- Volume name
- Container using the volume
- Mount path
- Size

Use arrow keys to navigate, Space to toggle selection, and Enter to confirm.

### Multiple Containers

Migrate volumes from multiple containers:

```bash
volume-migrator web-app db-server cache --remote user@production.example.com -i
```

### Custom SSH Key

Specify a custom SSH private key:

```bash
volume-migrator app --remote user@host --ssh-key ~/.ssh/deploy_key
```

### Dry Run

See what would be migrated without actually doing it:

```bash
volume-migrator app --remote user@host --verbose --dry-run
```

### Verbose Output

Show detailed progress information:

```bash
volume-migrator app --remote user@host --verbose
```

### Force Mode

Skip disk space validation checks:

```bash
volume-migrator app --remote user@host --force
```

### Configuration Validation

Validate configuration before running:

```bash
volume-migrator app --remote user@host --validate-only
```

## Command-Line Options

```
Flags:
  -r, --remote string                  Remote host in format user@host[:port] (required)
  -i, --interactive                    Display volumes and let user select which to migrate
      --ssh-key string                 Path to SSH private key (default: auto-detect)
      --ssh-port string                SSH port (default "22")
      --temp-dir string                Local temporary directory (default: /tmp/volume-migration-{timestamp})
      --remote-temp-dir string         Remote temporary directory (default: /tmp/volume-migration-{timestamp})
  -v, --verbose                        Verbose output
      --dry-run                        Show what would be done without doing it
      --validate-only                  Validate configuration without running migration
      --force                          Skip disk space validation checks
      --no-cleanup                     Keep temporary files for debugging
  -p, --progress                       Show progress bars during transfer (default true)
      --strict-host-key-checking       Verify SSH host keys against known_hosts (default true)
      --accept-host-key                Automatically accept and add unknown host keys (DANGEROUS - use only in trusted environments)
      --known-hosts-file string        Path to known_hosts file (default: ~/.ssh/known_hosts)
  -h, --help                           Help for volume-migrator

Commands:
  version     Print version information
```

## Security Best Practices

### SSH Host Key Verification

By default, the tool **verifies SSH host keys** against your `~/.ssh/known_hosts` file to prevent MITM attacks. This is enabled with `--strict-host-key-checking` (default: true).

**RECOMMENDED**: For first-time connections, manually verify the host key:

```bash
# Add the host to known_hosts first
ssh-keyscan remote-host >> ~/.ssh/known_hosts

# Or connect manually once
ssh user@remote-host

# Then run the migrator
volume-migrator mycontainer --remote user@remote-host
```

**ONLY FOR TRUSTED ENVIRONMENTS**: You can auto-accept unknown keys (not recommended for production):

```bash
volume-migrator mycontainer --remote user@host --accept-host-key
```

### SSH Key Permissions

Ensure proper permissions on SSH keys:

```bash
chmod 700 ~/.ssh
chmod 600 ~/.ssh/id_rsa
chmod 644 ~/.ssh/id_rsa.pub
chmod 644 ~/.ssh/known_hosts
```

### Docker Socket Security

When running in Docker, the container needs access to the Docker socket. Be aware:

- Mounting `/var/run/docker.sock` gives root-equivalent access
- Only run the container in trusted environments
- Consider using Docker contexts instead of socket mounting

### Configuration Validation

Always validate configuration before running in production:

```bash
# Validate configuration
volume-migrator mycontainer --remote user@host --validate-only

# Dry-run to see what would happen
volume-migrator mycontainer --remote user@host --dry-run --verbose
```

### Disk Space Checks

The tool validates disk space before migration by default. Use `--force` only when:
- You've manually verified sufficient space exists
- The estimation is incorrect for your use case
- You're testing or debugging

### Secrets Management

- Never commit SSH keys to version control
- Use SSH agent forwarding when possible
- Consider using secret management tools (Vault, AWS Secrets Manager)
- Rotate SSH keys regularly

## How It Works

1. **Initialization**: Connects to local Docker and detects sudo requirements
2. **SSH Connection**: Establishes secure connection to remote host with host key verification
3. **Volume Discovery**: Inspects specified containers and extracts volume information
4. **Disk Space Validation**: Checks available space on local and remote machines
5. **Selection** (if interactive): User selects which volumes to migrate
6. **Export**: Creates tar.gz archives of selected volumes using Alpine containers
7. **Transfer**: Uploads archives to remote host via SFTP with progress tracking
8. **Import**: Creates volumes on remote and extracts archive data
9. **Cleanup**: Removes temporary files on both local and remote machines

## SSH Authentication

The tool tries authentication methods in this order:

1. **SSH Agent** (if `SSH_AUTH_SOCK` is set)
2. **Custom key** (if `--ssh-key` is specified)
3. **Common private keys** (~/.ssh/id_rsa, id_ed25519, id_ecdsa)
4. **Password prompt** (fallback)

## Testing

### Run Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# View coverage report
go tool cover -html=coverage.out
```

### Test Coverage

Current test coverage: **40%+**

Covered areas:
- SSH authentication and host key verification
- Docker volume operations
- Configuration validation
- Logging utilities
- Byte size formatting and parsing

## Using Docker

### Build Docker Image

```bash
docker build -t volume-migrator:latest .
```

### Run with Docker

```bash
# Basic usage
docker run --rm \
  -v /var/run/docker.sock:/var/run/docker.sock:ro \
  -v ~/.ssh:/home/migrator/.ssh:ro \
  volume-migrator:latest mycontainer --remote user@host --dry-run

# With custom SSH key
docker run --rm \
  -v /var/run/docker.sock:/var/run/docker.sock:ro \
  -v ~/.ssh:/root/.ssh:ro \
  -v /path/to/key:/keys/deploy_key:ro \
  volume-migrator:latest mycontainer --remote user@host --ssh-key /keys/deploy_key

# Version command
docker run --rm volume-migrator:latest version
```

### Using Docker Compose

See `docker-compose.yml` for a complete example:

```bash
# Build
docker-compose build

# Run
docker-compose run --rm volume-migrator mycontainer --remote user@host --verbose

# Version
docker-compose run --rm volume-migrator version
```

## Examples

### Example 1: Simple Migration

```bash
# Create a test container with a volume
docker run -d --name test-app -v app-data:/data alpine sleep 3600
docker exec test-app sh -c "echo 'Hello World' > /data/test.txt"

# Migrate the volume
volume-migrator test-app --remote user@remote-server --verbose
```

### Example 2: Interactive Selection

```bash
# Multiple containers with shared volumes
docker run -d --name web -v shared-data:/app/data alpine sleep 3600
docker run -d --name worker -v shared-data:/app/data -v worker-logs:/logs alpine sleep 3600

# Interactively select volumes
volume-migrator web worker --remote user@remote-server --interactive
```

### Example 3: Production Migration with Security

```bash
# 1. Validate configuration first
volume-migrator webapp database redis \
  --remote deploy@production.example.com \
  --ssh-key ~/.ssh/production_deploy_key \
  --validate-only

# 2. Dry run to see what would happen
volume-migrator webapp database redis \
  --remote deploy@production.example.com \
  --ssh-key ~/.ssh/production_deploy_key \
  --dry-run \
  --verbose

# 3. Actual migration
volume-migrator webapp database redis \
  --remote deploy@production.example.com \
  --ssh-key ~/.ssh/production_deploy_key \
  --verbose \
  --progress
```

## Verification

After migration, verify on the remote host:

```bash
# SSH to remote host
ssh user@remote-host

# List volumes
docker volume ls

# Inspect volume data
docker run --rm -v <volume-name>:/data alpine ls -la /data
docker run --rm -v <volume-name>:/data alpine cat /data/<file>
```

## Troubleshooting

### Docker Not Accessible

If you get "docker is not accessible" error:

- Ensure Docker is installed: `docker --version`
- Check permissions: Try `sudo docker ps`
- If sudo is required, the tool will detect and use it automatically

### SSH Connection Failed

- Verify SSH access: `ssh user@remote-host`
- Check SSH key permissions: `chmod 600 ~/.ssh/id_rsa`
- Use `--ssh-key` to specify correct key
- Ensure SSH agent is running: `eval $(ssh-agent) && ssh-add`
- Check host key verification: `ssh-keyscan remote-host >> ~/.ssh/known_hosts`

### Configuration Validation Failed

Use `--validate-only` to see what's wrong:

```bash
volume-migrator mycontainer --remote user@host --validate-only
```

Common issues:
- Invalid remote host format (must be `user@host` or `user@host:port`)
- Conflicting flags (`--strict-host-key-checking` and `--accept-host-key`)
- SSH key file doesn't exist
- Invalid SSH port number

### Volume Not Found

- Verify container exists: `docker ps -a | grep container-name`
- Check container has volumes: `docker inspect container-name`
- Only named volumes are migrated (bind mounts are skipped)

### Insufficient Disk Space

The tool checks disk space before migration. If you get this error:

- Free up space on local or remote machine
- Use `--force` to skip validation (not recommended)
- Check actual volume sizes: `docker system df -v`

### Transfer Failed

- Check network connectivity to remote host
- Verify sufficient disk space on both machines
- Check remote temp directory permissions
- Review verbose logs: `--verbose`

### Remote Docker Issues

- Ensure Docker is installed on remote: `ssh user@host docker --version`
- The tool auto-detects sudo requirements on remote
- Verify user has Docker permissions or sudo access

## Limitations

- Only migrates **named volumes** (bind mounts are not supported)
- Both local and remote must have the Alpine Docker image available
- Requires SSH access to remote machine
- Large volumes may take significant time to transfer
- Disk space estimation assumes ~67% compression ratio (tar.gz)

## Development

### Running Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run linter
make lint

# Run go vet
make vet
```

### Building

```bash
# Build for current platform
make build

# Build for Linux
make build-linux

# Build for all platforms
make build-all

# Install locally
make install

# Show version info
make version

# Show help
make help
```

### Project Structure

```
VolumeMigration/
├── .github/
│   └── workflows/           # GitHub Actions CI/CD
│       ├── ci.yml          # Continuous Integration
│       └── release.yml     # Automated releases
├── cmd/volume-migrator/    # Main CLI application
├── internal/
│   ├── docker/             # Docker client and operations
│   ├── ssh/                # SSH client and SFTP transfer
│   ├── migrator/           # Migration orchestration
│   ├── ui/                 # Interactive UI components
│   ├── utils/              # Logging and utilities
│   └── errors/             # Custom error types
├── Dockerfile              # Container image
├── docker-compose.yml      # Docker Compose example
├── .dockerignore           # Docker build exclusions
├── .gitignore              # Git exclusions
├── .golangci.yml           # Linter configuration
├── go.mod
├── go.sum
├── Makefile
├── LICENSE
├── README.md
└── TODO.md
```

## CI/CD

The project uses GitHub Actions for continuous integration and deployment:

- **CI Workflow**: Runs on every push and PR
  - Tests on Ubuntu, macOS, and Windows
  - Tests with Go 1.21, 1.22, and 1.23
  - Runs linter and go vet
  - Builds for all platforms
  - Uploads coverage to Codecov

- **Release Workflow**: Runs on version tags (v*)
  - Builds binaries for all platforms
  - Creates tar.gz/zip archives
  - Generates SHA256 checksums
  - Creates GitHub releases automatically

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run tests and linter: `make test lint`
6. Submit a pull request

## License

MIT License - see [LICENSE](LICENSE) file for details

## Author

Created with Claude Code

## Acknowledgments

- Built with [Cobra](https://github.com/spf13/cobra) for CLI
- Uses [logrus](https://github.com/sirupsen/logrus) for structured logging
- SSH operations powered by [golang.org/x/crypto/ssh](https://pkg.go.dev/golang.org/x/crypto/ssh)
- Interactive UI with [go-prompt](https://github.com/c-bata/go-prompt)
