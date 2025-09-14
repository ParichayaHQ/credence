---
layout: guide
title: "Light Node Guide"
description: "Run a lightweight Credence node for network participation"
collection: guides
status: beta
---

# Light Node Deployment Guide

> **⚠️ IMPLEMENTATION STATUS**  
> The core services (p2p-gateway, fullnode) exist in the codebase. However, the **Docker containers, docker-compose files, configuration templates, deployment scripts, and release packages** described here are not yet implemented. This guide represents the target deployment experience.

## Overview

A light node provides basic network infrastructure by participating in P2P gossip and maintaining recent data. Perfect for users who want to contribute to network health without running full services.

## Requirements

### System Requirements
- **CPU**: 2 cores minimum
- **RAM**: 4GB minimum  
- **Storage**: 20GB available space
- **Network**: Stable internet connection, ports 4001/tcp open
- **OS**: Linux, macOS, or Windows with Docker

### Dependencies
- Docker 20.10+
- Docker Compose 2.0+
- 100 Mbps+ internet connection

## Quick Start

### 1. Download Configuration
```bash
curl -L https://github.com/ParichayaHQ/credence/releases/latest/download/light-node.tar.gz -o light-node.tar.gz
tar -xzf light-node.tar.gz
cd light-node/
```

### 2. Configure Node
```bash
# Copy example configuration
cp config/light-node.example.yml config/light-node.yml

# Edit configuration (see Configuration section below)
nano config/light-node.yml
```

### 3. Start Services
```bash
# Start light node stack
docker-compose -f docker-compose.light.yml up -d

# Check status
docker-compose -f docker-compose.light.yml ps
```

### 4. Verify Operation
```bash
# Check P2P connectivity
curl http://localhost:8080/health

# View logs
docker-compose -f docker-compose.light.yml logs -f
```

## Services Included

### P2P Gateway
- **Gossipsub participation** - Relays events across network
- **DHT participation** - Helps with peer discovery
- **Rate limiting** - Protects against spam/DoS
- **HTTP bridge** - Provides local API for other services

### Full Node (Light Mode)
- **Event storage** - Keeps recent events (30 days default)
- **Query API** - Serves data to wallets and other nodes
- **Sync protocol** - Maintains consistency with network
- **Pruning** - Automatically removes old data to save space

## Configuration

### Light Node Config (`config/light-node.yml`)
```yaml
# Network Configuration
network:
  listen_addresses:
    - "/ip4/0.0.0.0/tcp/4001"
  bootstrap_peers:
    - "/dns4/seed1.credence.network/tcp/4001/p2p/12D3..."
    - "/dns4/seed2.credence.network/tcp/4001/p2p/12D4..."
  
# Storage Configuration  
storage:
  retention_days: 30        # Keep data for 30 days
  max_storage_gb: 10        # Limit storage usage
  prune_interval: "24h"     # Check for old data daily

# API Configuration
api:
  listen: "0.0.0.0:8080"
  cors_origins: ["*"]
  rate_limit: "100/minute"

# Resource Limits
resources:
  max_memory: "2GB"
  max_cpu: "1.0"
  max_connections: 100
```

### Docker Compose (`docker-compose.light.yml`)
```yaml
version: '3.8'

services:
  p2p-gateway:
    image: credence/p2p-gateway:latest
    ports:
      - "4001:4001"
      - "8080:8080"
    volumes:
      - ./config:/config
      - ./data/p2p:/data
    environment:
      - CONFIG_PATH=/config/light-node.yml
      - LOG_LEVEL=info
    restart: unless-stopped

  fullnode:
    image: credence/fullnode:latest
    ports:
      - "8081:8081"
    volumes:
      - ./config:/config
      - ./data/fullnode:/data
    depends_on:
      - p2p-gateway
    environment:
      - CONFIG_PATH=/config/light-node.yml
      - MODE=light
      - P2P_ENDPOINT=http://p2p-gateway:8080
    restart: unless-stopped

volumes:
  p2p-data:
  fullnode-data:
```

## Management

### Starting/Stopping
```bash
# Start services
docker-compose -f docker-compose.light.yml up -d

# Stop services
docker-compose -f docker-compose.light.yml down

# Restart services
docker-compose -f docker-compose.light.yml restart

# View status
docker-compose -f docker-compose.light.yml ps
```

### Monitoring
```bash
# View logs
docker-compose -f docker-compose.light.yml logs -f

# Check resource usage
docker stats

# Monitor P2P connectivity
curl http://localhost:8080/peers | jq .

# Check data sync status
curl http://localhost:8081/sync/status | jq .
```

### Updates
```bash
# Pull latest images
docker-compose -f docker-compose.light.yml pull

# Restart with new images
docker-compose -f docker-compose.light.yml up -d
```

## Networking

### Firewall Configuration
```bash
# Allow P2P traffic
sudo ufw allow 4001/tcp comment "Credence P2P"

# Allow API access (optional, for local development)
sudo ufw allow from 192.168.0.0/16 to any port 8080
sudo ufw allow from 192.168.0.0/16 to any port 8081
```

### NAT/Router Setup
- **Port forwarding**: Forward external port 4001 to your node
- **UPnP**: Enable if available for automatic port mapping
- **Static IP**: Recommended for consistent peer connectivity

## Troubleshooting

### Common Issues

**No peers connecting:**
```bash
# Check firewall
sudo ufw status

# Test external connectivity
nc -zv your-external-ip 4001

# Verify bootstrap peers
curl http://localhost:8080/peers
```

**High resource usage:**
```bash
# Adjust resource limits in config
nano config/light-node.yml

# Restart with new limits
docker-compose -f docker-compose.light.yml restart
```

**Storage filling up:**
```bash
# Check current usage
du -sh data/

# Adjust retention period
nano config/light-node.yml  # Reduce retention_days

# Force pruning
docker-compose -f docker-compose.light.yml exec fullnode /app/prune-now
```

### Health Checks
```bash
# P2P Gateway health
curl http://localhost:8080/health
# Should return: {"status": "healthy", "peers": N, "uptime": "Xs"}

# Full Node health  
curl http://localhost:8081/health
# Should return: {"status": "healthy", "events": N, "sync": "up-to-date"}
```

## Economics

### Costs
- **Bandwidth**: ~10-50 GB/month
- **Storage**: ~5-15 GB
- **CPU**: Low utilization
- **Power**: ~5-15W additional usage

### Rewards
- **Network tokens** earned for providing service
- **Proportional to uptime** and data served
- **Paid out monthly** via on-chain transactions

## Scaling Up

Ready for more participation? Consider upgrading to:
- **[Full Node](./full-node.md)** - Complete services and higher rewards
- **[Validator Node](./validator-node.md)** - Participate in consensus
- **[Service Provider](./service-provider.md)** - Specialized service offerings

Running a light node helps make Credence more decentralized and resilient!