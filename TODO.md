# VolumeMigration - Todo List

> Prioritized list of improvements, bugs, and enhancements for the VolumeMigration project.

**Last Updated**: 2026-01-07
**Current Readiness Score**: 85/100

---

## 🔴 Critical (Security & Correctness) - MUST DO

These are **blocking issues** that must be addressed before production use.

### 1. Fix SSH Host Key Verification [BLOCKING]

**Priority**: CRITICAL
**Difficulty**: Medium
**Files**: `internal/ssh/client.go`
**Line**: 36

**Problem**: Currently uses `ssh.InsecureIgnoreHostKey()` which accepts any SSH host key without verification, making the connection vulnerable to man-in-the-middle (MITM) attacks.

**Tasks**:
- [ ] Remove `ssh.InsecureIgnoreHostKey()` from `internal/ssh/client.go:36`
- [ ] Implement known_hosts file reading (read from `~/.ssh/known_hosts`)
- [ ] Add `--strict-host-key-checking` flag (default: true)
- [ ] Add confirmation prompt for unknown host keys in interactive mode
- [ ] Support `--accept-host-key` flag for automation (with clear warning)
- [ ] Add host key fingerprint display in verbose mode
- [ ] Update README with security best practices section
- [ ] Add tests for host key verification logic

**Recommended Approach**:
```go
// Option 1: Use known_hosts
hostKeyCallback, err := knownhosts.New(filepath.Join(os.Getenv("HOME"), ".ssh", "known_hosts"))

// Option 2: For automation, use FixedHostKey with provided fingerprint
hostKeyCallback := ssh.FixedHostKey(publicKey)
```

---

### 2. Generate and Commit go.sum [BLOCKING]

**Priority**: CRITICAL
**Difficulty**: Trivial
**Files**: Root directory

**Problem**: `go.mod` exists but `go.sum` is missing, which prevents dependency integrity verification and breaks reproducible builds.

**Tasks**:
- [ ] Run `go mod download` to fetch dependencies
- [ ] Run `go mod tidy` to clean up and generate go.sum
- [ ] Verify go.sum is created
- [ ] Commit go.sum to repository
- [ ] Add go.sum validation to CI pipeline (when created)
- [ ] Document dependency management in README

**Commands**:
```bash
go mod download
go mod tidy
git add go.sum
git commit -m "Add go.sum for dependency integrity"
```

---

### 3. Add Test Coverage [HIGH]

**Priority**: HIGH
**Difficulty**: High
**Impact**: No automated quality assurance currently

**Problem**: Zero test coverage despite Makefile having a `test` target. No way to catch regressions or verify functionality.

**Target**: Minimum 40% code coverage for critical paths

**Tasks**:

#### 3.1 Docker Package Tests
- [ ] Create `internal/docker/sudo_test.go`
  - Test sudo detection with mocked Docker commands
  - Test sudo requirement caching
  - Test thread-safety of sudo detector

- [ ] Create `internal/docker/volume_test.go`
  - Test size parsing with fixtures (KB, MB, GB, TB, edge cases)
  - Test volume discovery from mock container inspect output
  - Test volume deduplication logic
  - Test bind mount filtering

- [ ] Create `internal/docker/client_test.go`
  - Test container inspection parsing
  - Test command building with sudo wrapper

#### 3.2 SSH Package Tests
- [ ] Create `internal/ssh/auth_test.go`
  - Test host string parsing (user@host:port variations)
  - Test SSH key file discovery
  - Test authentication method priority

- [ ] Create `internal/ssh/client_test.go`
  - Test command execution (mocked sessions)
  - Test remote sudo detection

#### 3.3 Migrator Package Tests
- [ ] Create `internal/migrator/export_test.go`
  - Test archive path generation
  - Test export command building
  - Test cleanup on export failure

- [ ] Create `internal/migrator/import_test.go`
  - Test import command building
  - Test error handling and cleanup

#### 3.4 Test Infrastructure
- [ ] Set up test fixtures directory
- [ ] Create mock Docker client
- [ ] Create mock SSH client
- [ ] Add table-driven tests for parsing functions
- [ ] Add integration test script (requires Docker)
- [ ] Add coverage reporting to CI

**Resources Needed**:
- Mock library for exec.Command (e.g., `github.com/golang/mock`)
- Test fixtures for Docker inspect output
- Test fixtures for docker system df output

---

## 🟡 High Priority (Code Quality)

These improvements significantly enhance reliability and maintainability.

### 4. Improve Error Handling

**Priority**: MEDIUM
**Difficulty**: Low
**Files**: Multiple, especially `internal/migrator/import.go`

**Tasks**:

- [ ] Fix swallowed cleanup errors in `internal/migrator/import.go:40-42`
  - Log cleanup failures separately
  - Don't ignore cleanup errors (at least log them properly)

- [ ] Add error wrapping with context throughout codebase
  - Ensure all errors include operation context
  - Use `fmt.Errorf("operation failed for volume %s: %w", name, err)`

- [ ] Create custom error types for common failures
  - `VolumeNotFoundError`
  - `SSHConnectionError`
  - `DiskSpaceError`
  - `PermissionError`

- [ ] Add error recovery for partial multi-volume migrations
  - Track which volumes succeeded
  - Provide summary of successes/failures
  - Add `--continue-on-error` flag option

---

### 5. Implement Structured Logging

**Priority**: MEDIUM
**Difficulty**: Medium
**Files**: All packages (currently using fmt.Printf)

**Problem**: Logrus is imported but barely used. Most logging is via `fmt.Printf`, making it hard to filter, format, or aggregate logs.

**Tasks**:

- [ ] Replace all `fmt.Printf` with logrus throughout codebase
  - `fmt.Printf` → `log.Info()`
  - Errors → `log.WithError(err).Error()`

- [ ] Add structured logging fields
  - `log.WithField("volume", volName).Info("Exporting volume")`
  - `log.WithFields(log.Fields{"container": name, "host": host})`

- [ ] Implement proper log levels
  - DEBUG: Verbose operations, command output
  - INFO: Major steps (export started, transfer complete)
  - WARN: Non-fatal issues (cleanup failed but continuing)
  - ERROR: Fatal errors that stop migration

- [ ] Make log format configurable
  - Add `--log-format` flag (text/json)
  - JSON format for log aggregation systems
  - Text format for human readability

- [ ] Remove or integrate unused `internal/utils/logger.go`
  - Either use it consistently or remove it

- [ ] Add log file output option
  - Add `--log-file` flag for persistent logging
  - Useful for debugging and auditing

---

### 6. Add Configuration Validation

**Priority**: MEDIUM
**Difficulty**: Low
**Files**: `internal/migrator/migrator.go`

**Problem**: Minimal validation at startup. Invalid configs can cause cryptic errors deep in execution.

**Tasks**:

- [ ] Validate remote host format in `NewMigrator()`
  - Regex check for `user@host` or `user@host:port`
  - Reject invalid formats early with clear message

- [ ] Validate SSH port is numeric and in valid range
  - Check port is 1-65535
  - Default to 22 if not specified

- [ ] Validate temp directory paths
  - Check paths are absolute
  - Check paths are writable (local)
  - Create directories if they don't exist

- [ ] Validate container names
  - Check containers are non-empty strings
  - Check containers exist before starting migration
  - Provide clear error if container not found

- [ ] Add early validation phase
  - Fail fast with clear error messages
  - Run all validations before starting export
  - Add `--validate-only` flag for dry-run config check

---

### 7. Add Disk Space Validation

**Priority**: MEDIUM
**Difficulty**: Medium
**Files**: `internal/migrator/export.go`, `internal/ssh/transfer.go`

**Problem**: No check before export or transfer. Can fail mid-operation when disk fills up.

**Tasks**:

- [ ] Check local disk space before export
  - Calculate total size of volumes to export
  - Check available space in temp directory
  - Require at least 1.5x volume size (for compression overhead)

- [ ] Check remote disk space before transfer
  - Run `df` command via SSH on remote temp directory
  - Verify sufficient space for transfer + extraction
  - Account for compression ratio (~2x uncompressed size)

- [ ] Add `--force` flag to override space checks
  - Useful when space calculation is incorrect
  - Show warning when forcing

- [ ] Show space requirements in dry-run mode
  - Display calculated space needed
  - Display available space
  - Show pass/fail for space check

- [ ] Handle disk full errors gracefully
  - Detect "no space left on device" errors
  - Cleanup partial files
  - Provide clear error message with space info

---

## 🔵 Medium Priority (DevOps & Tooling)

These improvements enhance development workflow and release process.

### 8. Add GitHub Actions CI/CD

**Priority**: HIGH
**Difficulty**: Medium
**Files**: Create `.github/workflows/ci.yml`

**Tasks**:

- [ ] Create `.github/workflows/ci.yml` workflow file

- [ ] Add Go test job
  - Run on Ubuntu, macOS, Windows
  - Test against Go 1.21, 1.22, 1.23
  - Run `go test -v -race -coverprofile=coverage.out ./...`

- [ ] Add Go vet job
  - Run `go vet ./...`
  - Fail on any vet issues

- [ ] Add golangci-lint job
  - Install golangci-lint
  - Run `golangci-lint run`
  - Cache linter results

- [ ] Add build job
  - Build for linux/amd64
  - Build for darwin/amd64
  - Build for darwin/arm64 (Apple Silicon)
  - Build for windows/amd64

- [ ] Upload build artifacts
  - Store binaries as GitHub Actions artifacts
  - Include checksums (SHA256)

- [ ] Add status badge to README
  - `[![CI](https://github.com/user/repo/workflows/CI/badge.svg)](https://github.com/user/repo/actions)`

**Example Workflow**:
```yaml
name: CI

on: [push, pull_request]

jobs:
  test:
    strategy:
      matrix:
        go-version: [1.21, 1.22, 1.23]
        os: [ubuntu-latest, macos-latest, windows-latest]
    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: ${{ matrix.go-version }}
      - run: go test -v -race -coverprofile=coverage.out ./...
      - run: go vet ./...
```

---

### 9. Add Linting Configuration

**Priority**: MEDIUM
**Difficulty**: Low
**Files**: Create `.golangci.yml`

**Tasks**:

- [ ] Create `.golangci.yml` configuration file

- [ ] Enable standard linters
  - govet (bug detection)
  - errcheck (unchecked errors)
  - staticcheck (static analysis)
  - gosimple (simplification suggestions)
  - unused (unused code detection)
  - ineffassign (ineffective assignments)

- [ ] Configure linter settings
  - Set line length limit (120 chars)
  - Enable/disable specific checks
  - Configure exclusions for generated code

- [ ] Fix existing linting issues
  - Run `golangci-lint run` locally
  - Fix all reported issues
  - Commit fixes

- [ ] Integrate into CI pipeline
  - Add linting job to GitHub Actions
  - Fail builds on lint errors

- [ ] Add pre-commit hook (optional)
  - Run linting before commit
  - Prevent committing code with lint errors

---

### 10. Add .gitignore

**Priority**: LOW
**Difficulty**: Trivial
**Files**: Create `.gitignore`

**Tasks**:

- [ ] Create `.gitignore` file with appropriate exclusions
- [ ] Check if `bin/` directory was accidentally committed (remove if so)
- [ ] Verify `.gitignore` works correctly

**Recommended Content**:
```
# Binaries
bin/
*.exe
*.dll
*.so
*.dylib

# Test coverage
*.out
coverage.html

# Temporary files
*.tar.gz
*.tmp
.DS_Store

# IDE
.vscode/
.idea/
*.swp
*.swo
*~

# Environment
.env
.env.local
```

---

### 11. Add Release Automation

**Priority**: MEDIUM
**Difficulty**: Medium
**Files**: Create `.github/workflows/release.yml`

**Tasks**:

- [ ] Create `.github/workflows/release.yml` triggered on tags

- [ ] Build for multiple architectures
  - linux/amd64
  - linux/arm64
  - darwin/amd64 (Intel Mac)
  - darwin/arm64 (Apple Silicon)
  - windows/amd64

- [ ] Generate checksums
  - Create SHA256 checksums for all binaries
  - Include checksums.txt in release

- [ ] Auto-create GitHub releases
  - Extract version from git tag (v1.0.0)
  - Generate release notes from commits
  - Upload all binaries and checksums

- [ ] Add installation instructions
  - Update README with installation from releases
  - Include checksum verification steps

**Example Release Workflow**:
```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  build:
    strategy:
      matrix:
        include:
          - os: ubuntu-latest
            goos: linux
            goarch: amd64
          - os: ubuntu-latest
            goos: darwin
            goarch: amd64
          - os: ubuntu-latest
            goos: windows
            goarch: amd64
```

---

### 12. Add Version Information

**Priority**: LOW
**Difficulty**: Low
**Files**: `cmd/volume-migrator/main.go`

**Tasks**:

- [ ] Add `--version` flag to CLI

- [ ] Inject version via ldflags during build
  - Update Makefile to use ldflags
  - Pass version, commit hash, build date

- [ ] Display version information
  - `volume-migrator --version` shows full version info
  - Include in verbose mode output

- [ ] Update build commands

**Example Makefile change**:
```makefile
VERSION := $(shell git describe --tags --always --dirty)
COMMIT := $(shell git rev-parse --short HEAD)
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

build:
	go build -ldflags "$(LDFLAGS)" -o bin/volume-migrator ./cmd/volume-migrator
```

**Example version output**:
```
volume-migrator version v1.2.3
Commit: abc1234
Built: 2026-01-07T10:30:00Z
Go version: go1.21.5
```

---

## 🟢 Low Priority (Enhancements)

These are nice-to-have features and optimizations.

### 13. Implement SFTP Retry Logic

**Priority**: MEDIUM
**Difficulty**: High
**Files**: `internal/ssh/transfer.go`

**Tasks**:

- [ ] Implement exponential backoff for retries
  - First retry: 1s delay
  - Second retry: 2s delay
  - Third retry: 4s delay
  - Max retries: 5

- [ ] Make retry count configurable
  - Add `--retry-count` flag
  - Add `--retry-delay` flag for initial delay

- [ ] Add resume capability for partial transfers
  - Track bytes transferred
  - Resume from last position on retry
  - Verify partial file integrity

- [ ] Log retry attempts
  - Log each retry with attempt number
  - Log final success or failure

- [ ] Add retry for temporary network failures
  - Detect specific network errors (connection reset, timeout)
  - Don't retry on permanent failures (permission denied)

---

### 14. Add Concurrent Processing

**Priority**: LOW
**Difficulty**: High
**Files**: `internal/migrator/migrator.go`

**Problem**: Sequential processing only. Large migrations take a long time.

**Tasks**:

- [ ] Add worker pool for volume export
  - Export multiple volumes concurrently
  - Limit concurrency to prevent resource exhaustion

- [ ] Pipeline export and transfer
  - Start transferring while still exporting other volumes
  - Use channels to coordinate

- [ ] Add `--concurrency` flag
  - Control number of concurrent operations
  - Default to 3

- [ ] Handle errors in concurrent operations
  - Track partial failures
  - Cancel remaining operations on critical error
  - Provide summary of successes/failures

- [ ] Add progress tracking for concurrent operations
  - Show overall progress across all volumes
  - Show individual volume progress

---

### 15. Additional Features

**Priority**: LOW
**Difficulty**: Varies

#### 15.1 Password Authentication Fallback
- [ ] Implement password auth (currently not supported despite README)
- [ ] Add keyboard-interactive auth method
- [ ] Add password prompt with masking
- [ ] **Files**: `internal/ssh/auth.go`

#### 15.2 Auto-Pull Alpine Image
- [ ] Detect missing Alpine image
- [ ] Offer to pull image (with confirmation)
- [ ] Show pull progress
- [ ] **Files**: `internal/docker/client.go`

#### 15.3 Progress for Export/Import Operations
- [ ] Show spinner during tar operations
- [ ] Estimate time based on volume size
- [ ] Show current operation phase
- [ ] **Files**: `internal/migrator/export.go`, `internal/migrator/import.go`

#### 15.4 Compression Level Configuration
- [ ] Add `--compression-level` flag (1-9)
- [ ] Default to 6 (balance speed/size)
- [ ] Show compression ratio after export
- [ ] **Files**: `internal/migrator/export.go`

#### 15.5 Windows Support Improvements
- [ ] Use `os.TempDir()` instead of hardcoded `/tmp`
- [ ] Handle Windows path separators correctly
- [ ] Test on Windows
- [ ] **Files**: Multiple

#### 15.6 Containerized Execution
- [ ] Create Dockerfile for running volume-migrator in container
- [ ] Multi-stage build for minimal image size
- [ ] Mount Docker socket for local access
- [ ] Include SSH client dependencies
- [ ] **Files**: Create `Dockerfile`

---

## 📝 Documentation

### 16. Update README

**Priority**: MEDIUM
**Difficulty**: Low
**Files**: `README.md`

**Tasks**:

- [ ] Add security best practices section
  - Explain host key verification importance
  - Document --strict-host-key-checking flag (when implemented)
  - Recommend SSH key-based auth over passwords

- [ ] Add troubleshooting section for host key verification
  - How to add host to known_hosts
  - How to handle host key changes
  - How to use --accept-host-key for automation

- [ ] Document testing section
  - How to run tests
  - How to run with coverage
  - How to contribute tests

- [ ] Add CI badge
  - Link to GitHub Actions workflow
  - Show build status

- [ ] Add contributing guidelines
  - Code style requirements
  - How to submit PRs
  - Testing requirements

- [ ] Add installation from releases
  - Download from GitHub releases
  - Verify checksums
  - Install to PATH

---

## Summary by Priority

### Must Do (Blockers for Production)
1. ✅ Fix InsecureIgnoreHostKey vulnerability
2. ✅ Generate and commit go.sum
3. ✅ Add basic test coverage (minimum 40%)

### Should Do (Important for Quality)
4. ✅ Improve error handling
5. ✅ Implement structured logging
6. ✅ Add configuration validation
7. ✅ Add disk space validation
8. ✅ GitHub Actions CI/CD
9. ✅ Add linting configuration

### Nice to Have (Enhancements)
10. ✅ Release automation
11. ✅ Version information
12. ✅ SFTP retry logic
13. ✅ Concurrent processing
14. ✅ Additional features (password auth, auto-pull, etc.)

---

## Progress Tracking

**Critical Issues**: 0 / 3 completed
**High Priority**: 0 / 4 completed
**Medium Priority**: 0 / 5 completed
**Low Priority**: 0 / 2 completed

**Overall Completion**: 0 / 14 tasks (0%)

---

**Next Action**: Start with Critical tasks (Fix SSH host key verification, generate go.sum, add tests)
