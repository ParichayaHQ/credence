# Multi-stage build for fullnode service
FROM golang:1.24.1-alpine AS builder

# Install build dependencies  
RUN apk add --no-cache git make build-base

# Set working directory
WORKDIR /app

# Copy go mod files first (for better caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the fullnode binary with SQLite support (no RocksDB)
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o /app/bin/fullnode ./cmd/fullnode

# Final stage - minimal runtime image
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S credence && \
    adduser -u 1001 -S credence -G credence

# Create directories
RUN mkdir -p /data /config /shared-data && \
    chown -R credence:credence /data /config /shared-data

# Copy binary from builder stage
COPY --from=builder /app/bin/fullnode /usr/local/bin/fullnode

# Switch to non-root user
USER credence

# Expose ports
EXPOSE 8081

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8081/health || exit 1

# Set default environment variables
ENV LOG_LEVEL=info \
    API_PORT=8081 \
    DATA_DIR=/data \
    CONFIG_PATH=/config/fullnode.yml \
    MODE=full \
    P2P_ENDPOINT=http://p2p-gateway:8080

# Default command
CMD ["fullnode"]