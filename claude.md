# VolumeMigration - Project Guide for AI Assistants

> This document helps AI assistants (like Claude Code) understand the VolumeMigration codebase structure, patterns, and gotchas.

## Project Overview

**VolumeMigration** is a CLI tool for migrating Docker volumes from local containers to remote Linux machines.

**Purpose**: Automate the secure transfer of Docker volume data between systems without manual intervention or complex scripts.

**Primary Use Cases**:
- Production migrations when moving containers between servers
- Container relocations for infrastructure changes
- Backup and disaster recovery scenarios
- Development environment replication

**Key Value Propositions**:
- Single binary deployment (no external dependencies beyond Docker and SSH)
- Secure transfer via SSH/SFTP with progress tracking
- Interactive volume selection UI
- Automatic sudo detection (local and remote)
- Temporary file cleanup with debugging options

## Architecture Overview

### Package Responsibilities

```
VolumeMigration/
├── cmd/volume-migrator/          # CLI Application Layer
│   └── main.go                   # Entry point, Cobra CLI setup, flag parsing, signal handling
│
├── internal/
│   ├── migrator/                 # Orchestration Layer
│   │   ├── migrator.go          # Main workflow coordinator
│   │   ├── export.go            # Volume export to tar.gz archives
│   │   ├── import.go            # Volume import on remote host
│   │   └── cleanup.go           # Temporary file cleanup
│   │
│   ├── docker/                   # Docker Operations Layer
│   │   ├── client.go            # Docker client wrapper, container inspection
│   │   ├── sudo.go              # Auto-detection of sudo requirements
│   │   └── volume.go            # Volume discovery, size calculation
│   │
│   ├── ssh/                      # SSH/SFTP Operations Layer
│   │   ├── client.go            # SSH connection management
│   │   ├── auth.go              # SSH authentication (keys, agent)
│   │   └── transfer.go          # SFTP file transfer with progress
│   │
│   ├── ui/                       # User Interface Layer
│   │   └── selector.go          # Interactive volume selection
│   │
│   └── utils/                    # Shared Utilities
│       ├── logger.go            # Logging configuration
│       └── progress.go          # Progress bar utilities
```

### Key Design Patterns

1. **Facade Pattern**: `docker.Client` and `ssh.Client` wrap complex operations behind simple interfaces
2. **Strategy Pattern**: Multiple SSH authentication methods tried in priority order
3. **Template Method**: Migration workflow broken into distinct phases (discover → select → export → transfer → import → cleanup)
4. **Dependency Injection**: Context and config passed to constructors for testability

### Data Flow

```
User Input (CLI flags)
    ↓
Config struct construction
    ↓
Migrator initialization
    ├─→ Docker Client (local)
    └─→ SSH Client (remote)
    ↓
Volume discovery from containers
    ↓
Interactive selection (if -i flag)
    ↓
Export volumes to tar.gz (using Alpine containers)
    ↓
SFTP transfer to remote
    ↓
Import on remote (create volume + extract)
    ↓
Cleanup temporary files (unless --no-cleanup)
```

## Critical Files Map

| File | Importance | Purpose |
|------|-----------|---------|
| `cmd/volume-migrator/main.go` | **CRITICAL** | CLI entry point, flag definitions, command setup |
| `internal/migrator/migrator.go` | **CRITICAL** | Main orchestration logic, workflow coordination |
| `internal/docker/client.go` | **CRITICAL** | Docker operations, container inspection |
| `internal/docker/volume.go` | **CRITICAL** | Volume discovery, size calculation |
| `internal/ssh/client.go` | **CRITICAL** | SSH connection, remote command execution<br>⚠️ **HAS SECURITY ISSUE** at line 36 |
| `internal/ssh/transfer.go` | **CRITICAL** | SFTP file transfers with progress tracking |
| `internal/ssh/auth.go` | **IMPORTANT** | Multi-method SSH authentication |
| `internal/docker/sudo.go` | **IMPORTANT** | Sudo auto-detection for Docker |
| `internal/migrator/export.go` | **IMPORTANT** | Volume export to tar.gz archives |
| `internal/migrator/import.go` | **IMPORTANT** | Volume import on remote host |
| `internal/ui/selector.go` | Standard | Interactive volume selection UI |
| `internal/migrator/cleanup.go` | Standard | Temporary file cleanup |
| `internal/utils/*.go` | Standard | Logging and progress utilities |
| `go.mod` | **CRITICAL** | Dependencies ⚠️ **go.sum is MISSING** |

## Coding Conventions and Patterns

### Error Handling
- Always wrap errors with context: `fmt.Errorf("failed to export volume %s: %w", volName, err)`
- Check errors immediately after operations
- No custom error types currently used (consider adding for common failures)

### Context Usage
- Context passed through from main to all clients for cancellation support
- Used for graceful shutdown on Ctrl+C (SIGINT/SIGTERM)
- No timeout implementation currently (could be added)

### Docker Command Pattern
```go
// Always use sudo.WrapCommand() to conditionally add sudo
cmd := d.sudo.WrapCommand("docker", "volume", "create", volumeName)
output, err := exec.CommandContext(ctx, cmd[0], cmd[1:]...).CombinedOutput()
```

- Commands built as string slices for `exec.Command`
- Separate stdout/stderr buffers for error reporting
- Use `CombinedOutput()` for short commands, separate buffers for long-running

### SSH Command Pattern
```go
// General commands
output, err := client.RunCommand(ctx, "ls -la /tmp")

// Docker commands (automatically adds sudo if needed)
output, err := client.RunDockerCommand(ctx, "volume", "ls")
```

- New session created per command (not reused)
- Non-interactive execution (no PTY allocation)
- Command exit codes checked

### Volume Export/Import Technique
```go
// Export: Mount volume read-only into Alpine container
docker run --rm \
  -v volumeName:/data:ro \        // Read-only to prevent conflicts
  -v /tmp:/backup \               // Bind mount for output
  alpine tar czf /backup/vol.tar.gz -C /data .

// Import: Extract into newly created volume
docker run --rm \
  -v volumeName:/data \
  -v /tmp:/backup \
  alpine tar xzf /backup/vol.tar.gz -C /data
```

### Naming Conventions
- Exported types/functions: `PascalCase`
- Unexported: `camelCase`
- No interfaces currently defined (concrete types only)
- Package names are singular, lowercase

### Logging
- Logrus is imported but barely used (mostly `fmt.Printf`)
- Verbose flag controls detailed output
- No structured logging currently implemented
- **TODO**: Replace fmt.Printf with proper logrus usage

## Common Workflows

### Building the Project

```bash
# Build for current platform
make build
# Output: bin/volume-migrator

# Cross-compile for Linux
make build-linux
# Output: bin/volume-migrator-linux-amd64

# Install to GOPATH/bin
make install

# Clean build artifacts
make clean
```

### Running the Tool

```bash
# Basic usage (migrate all volumes from a container)
./bin/volume-migrator mycontainer --remote user@192.168.1.100

# Interactive mode (select which volumes to migrate)
./bin/volume-migrator mycontainer --remote user@host -i

# With custom SSH key
./bin/volume-migrator mycontainer --remote user@host --ssh-key ~/.ssh/custom_key

# Verbose output for debugging
./bin/volume-migrator mycontainer --remote user@host -v

# Dry-run (show what would be done without doing it)
./bin/volume-migrator mycontainer --remote user@host --dry-run

# Keep temp files for debugging (don't cleanup)
./bin/volume-migrator mycontainer --remote user@host --no-cleanup

# Multiple containers (migrate volumes from all)
./bin/volume-migrator container1 container2 container3 --remote user@host -i
```

### Dependency Management

```bash
# Download dependencies and generate go.sum (CURRENTLY MISSING!)
go mod download
go mod tidy

# Update dependencies
go get -u ./...
go mod tidy

# Vendor dependencies (optional)
go mod vendor
```

### Testing

```bash
# Run tests (currently no tests exist)
make test

# Run with coverage
go test -cover ./...

# Run with race detection
go test -race ./...
```

## CRITICAL Security & Gotchas

### 🔴 SECURITY ISSUE: Insecure SSH Host Key Verification

**Location**: `internal/ssh/client.go:36`

```go
HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: Implement proper host key verification
```

**Problem**: This accepts ANY SSH host key without verification, making the connection vulnerable to man-in-the-middle (MITM) attacks.

**Impact**: An attacker on the network could intercept the connection and steal credentials or data.

**Status**: TODO comment acknowledges this, but **MUST BE FIXED** before production use.

**Recommended Fix**:
1. Read from `~/.ssh/known_hosts`
2. Add `--strict-host-key-checking` flag (default: true)
3. Prompt user to confirm unknown host keys
4. Use `ssh.FixedHostKey()` for pre-shared fingerprints in automation scenarios

### ⚠️ Sudo Detection Gotchas

**Location**: `internal/docker/sudo.go`

- Auto-detection happens once at startup
- Uses `sudo -n` (non-interactive) which requires:
  - NOPASSWD configured in sudoers, OR
  - Cached sudo credentials from recent authentication
- If detection fails, entire tool fails (no fallback)
- Both local and remote sudo detected independently

**Potential Issues**:
- If sudo prompt appears, it will hang (non-interactive mode)
- Sudo credential cache timeout can cause failures mid-operation

### ⚠️ Missing go.sum File

**Location**: Root directory

**Problem**: `go.mod` exists but `go.sum` is missing from the repository

**Impact**:
- Cannot verify dependency integrity
- Breaks reproducible builds
- Security risk (dependencies not pinned)

**Fix**: Run `go mod download && go mod tidy` and commit the generated `go.sum`

### ⚠️ Zero Test Coverage

**Current State**: No `*_test.go` files exist despite Makefile having a `test` target

**Impact**:
- No automated quality assurance
- Bugs can't be caught before release
- Refactoring is risky

**Priority**: HIGH - Should add at least basic tests for critical paths

### ⚠️ Temp Directory Handling

- Default temp dirs use Unix-style paths (`/tmp`)
- Windows compatibility unclear (uses `filepath.Join` but hardcoded `/tmp`)
- Cleanup is deferred - if process crashes, temp files remain
- `--no-cleanup` useful for debugging but can fill disk

### ⚠️ Alpine Image Dependency

- Assumes `alpine` Docker image is available on both local and remote systems
- No automatic pull if missing (will fail with unclear error)
- No version pinning (implicitly uses `alpine:latest`)
- **Potential Issue**: Different Alpine versions between local/remote could cause issues

### ⚠️ Volume Limitations

- Only handles **named volumes** (bind mounts are automatically skipped)
- Volume should not be in active use during export (mounted read-only helps but not foolproof)
- No validation of available disk space before export/transfer (can fail mid-operation)
- Large volumes can take significant time with no granular progress for export/import phases

### ⚠️ Error Handling Gaps

**Location**: `internal/migrator/import.go:40-42`

```go
if err := m.cleanup(ctx); err != nil {
    // Error is logged but swallowed - cleanup failure is not surfaced to caller
}
```

- Partial failures in multi-volume migrations may leave inconsistent state
- Import failure attempts cleanup but errors are swallowed
- SFTP transfer errors don't provide retry logic (single attempt only)

### ⚠️ Concurrency Limitations

- No concurrent volume exports/transfers (sequential only)
- Context cancellation may leave partial state (e.g., volume created but not populated)
- No transaction-like rollback on failure

## Implementation Details

### SSH Authentication Priority Order

The tool tries authentication methods in this order:

1. **SSH Agent** (if `SSH_AUTH_SOCK` environment variable is set)
2. **Custom key** (if `--ssh-key` flag specified)
3. **Common keys** in order:
   - `~/.ssh/id_rsa`
   - `~/.ssh/id_ed25519`
   - `~/.ssh/id_ecdsa`
   - `~/.ssh/id_dsa`
4. **No password fallback** (despite README claiming it exists)

**Note**: Password authentication is NOT implemented despite documentation suggesting it is.

### Docker Volume Size Parsing

**Location**: `internal/docker/volume.go` - `parseSizeToBytes()` function

- Uses `docker system df -v` output
- Parses with regex: `^([\d.]+)([KMGT]?B?)$`
- Supports: B, KB, MB, GB, TB
- Falls back to "0B" if volume not found in df output
- Size is informational only (not validated before transfer)

### Container Inspection

**Location**: `internal/docker/client.go`

- Uses `docker inspect` with JSON parsing
- Only extracts volumes from Mounts array where `Type="volume"`
- Bind mounts are filtered out
- Container name automatically trimmed of leading `/`

### Volume Deduplication

**Location**: `internal/docker/volume.go` - `DiscoverVolumes()` function

- When multiple containers share a volume, it's only migrated once
- Uses map with volume name as key
- First container encountered "owns" the volume in metadata
- Prevents duplicate transfers

### Remote Command Execution

**Location**: `internal/ssh/client.go`

- Each SSH command creates a new session (not reused)
- No PTY allocation (non-interactive mode)
- Stdout and stderr captured separately
- Command exit codes checked for success
- Uses `session.CombinedOutput()` for short commands

### Progress Tracking

**Location**: `internal/utils/progress.go` and `internal/ssh/transfer.go`

- SFTP transfers show real-time progress based on file size
- Uses custom `ProgressReader` that wraps `io.Reader`
- Progress bars created with `github.com/schollz/progressbar/v3`
- Export/import operations don't show progress (tar runs inside container)
- Progress can be disabled with `--progress=false` flag

### Cross-Platform Considerations

- Built primarily for Linux target
- Uses `filepath.Join` for path construction
- Windows paths may conflict (temp dirs hardcoded as `/tmp` style)
- SSH key paths assume Unix-style home directories (`~/.ssh/`)
- Makefile includes `build-linux` target for cross-compilation

## Dependencies

**Core Libraries**:
- `github.com/spf13/cobra` - CLI framework
- `github.com/manifoldco/promptui` - Interactive terminal UI
- `github.com/sirupsen/logrus` - Structured logging (underutilized)
- `github.com/schollz/progressbar/v3` - Progress bars

**Network & Transfer**:
- `golang.org/x/crypto/ssh` - SSH protocol
- `github.com/pkg/sftp` - SFTP file transfer

**Build Requirements**:
- Go 1.21 or later
- Make (for build automation)
- Docker (for runtime)
- SSH client (for connections)

## Tips for AI Assistants

### When Adding Features
1. Follow existing patterns (Facade for clients, explicit error handling)
2. Add context parameter to all new functions
3. Use `sudo.WrapCommand()` for Docker operations
4. Check if operation needs sudo on remote via SSH client

### When Fixing Bugs
1. Add tests first (TDD approach)
2. Check both local and remote code paths
3. Verify cleanup happens even on errors
4. Test with and without sudo requirements

### When Refactoring
1. Don't break the single binary deployment model
2. Maintain backward compatibility with CLI flags
3. Keep packages focused (no circular dependencies)
4. Consider adding interfaces where helpful (currently missing)

### Critical Areas to Test
1. Sudo detection (local and remote)
2. SSH authentication fallback chain
3. Volume size parsing (various units)
4. Cleanup on errors
5. Context cancellation (Ctrl+C)
6. Multi-volume deduplication

### Before Proposing Changes
1. Check if change affects security (especially SSH code)
2. Verify change doesn't break cleanup logic
3. Consider impact on temp file handling
4. Test with both interactive and non-interactive modes
5. Verify both sudo and non-sudo paths

## Known Limitations

1. **No test coverage** - Manual testing only
2. **Insecure SSH** - MITM vulnerability
3. **Sequential processing** - No concurrent transfers
4. **No retry logic** - Single attempt for operations
5. **No disk space checks** - Can fail mid-transfer
6. **No progress for export/import** - Only SFTP shows progress
7. **Named volumes only** - Bind mounts not supported
8. **Alpine dependency** - Assumes image is available
9. **Unix-centric** - Windows compatibility unclear
10. **No transaction semantics** - Partial failures leave inconsistent state

## Next Steps for Contributors

See `TODO.md` for prioritized list of improvements, including:
- Fixing SSH host key verification (CRITICAL)
- Generating go.sum file (CRITICAL)
- Adding test coverage (HIGH)
- Implementing structured logging (HIGH)
- Adding CI/CD pipeline (MEDIUM)

---

**Last Updated**: 2026-01-07
**Codebase Version**: Based on analysis of current main branch
