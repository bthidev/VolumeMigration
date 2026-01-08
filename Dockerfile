# Multi-stage build for minimal image size

# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary with version info
ARG VERSION=dev
ARG COMMIT=none
ARG DATE=unknown

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags "-w -s -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}" \
    -o volume-migrator \
    ./cmd/volume-migrator

# Runtime stage
FROM alpine:3.19

# Install runtime dependencies
# - openssh-client: for SSH connections
# - docker-cli: for Docker operations
# - ca-certificates: for HTTPS
RUN apk add --no-cache \
    openssh-client \
    docker-cli \
    ca-certificates \
    tzdata

# Create non-root user
RUN addgroup -g 1000 migrator && \
    adduser -D -u 1000 -G migrator migrator

# Create necessary directories
RUN mkdir -p /home/migrator/.ssh && \
    chown -R migrator:migrator /home/migrator/.ssh && \
    chmod 700 /home/migrator/.ssh

# Copy binary from builder
COPY --from=builder /build/volume-migrator /usr/local/bin/volume-migrator

# Switch to non-root user
USER migrator

# Set working directory
WORKDIR /home/migrator

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD volume-migrator version || exit 1

# Default entrypoint
ENTRYPOINT ["volume-migrator"]

# Default command (show help)
CMD ["--help"]
