# Multi-stage build for lognode service (Trillian personality)
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

# Build the lognode binary
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o /app/bin/lognode ./cmd/lognode

# Final stage - minimal runtime image
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata postgresql-client

# Create non-root user
RUN addgroup -g 1001 -S credence && \
    adduser -u 1001 -S credence -G credence

# Create directories
RUN mkdir -p /data /config && \
    chown -R credence:credence /data /config

# Copy binary from builder stage
COPY --from=builder /app/bin/lognode /usr/local/bin/lognode

# Switch to non-root user
USER credence

# Expose ports
EXPOSE 8083

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8083/health || exit 1

# Set default environment variables
ENV LOG_LEVEL=info \
    API_PORT=8083 \
    DATA_DIR=/data \
    CONFIG_PATH=/config/lognode.yml \
    DATABASE_URL=postgres://credence:password@postgres/trillian \
    TREE_TYPE=merkle \
    SIGNING_FREQUENCY=1m

# Default command
CMD ["lognode"]