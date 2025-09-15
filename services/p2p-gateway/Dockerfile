# Multi-stage build for p2p-gateway service
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

# Build the p2p-gateway binary
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o /app/bin/p2p-gateway ./cmd/p2p-gateway

# Final stage - minimal runtime image
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S credence && \
    adduser -u 1001 -S credence -G credence

# Create directories
RUN mkdir -p /data /config && \
    chown -R credence:credence /data /config

# Copy binary from builder stage
COPY --from=builder /app/bin/p2p-gateway /usr/local/bin/p2p-gateway

# Switch to non-root user
USER credence

# Expose ports
EXPOSE 4001 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Set default environment variables
ENV LOG_LEVEL=info \
    P2P_PORT=4001 \
    API_PORT=8080 \
    DATA_DIR=/data \
    CONFIG_PATH=/config/p2p-gateway.yml

# Default command
CMD ["p2p-gateway"]