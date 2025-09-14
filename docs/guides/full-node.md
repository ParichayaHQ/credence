# Full Node Deployment Guide

> **⚠️ IMPLEMENTATION STATUS**  
> The core services (p2p-gateway, fullnode, scorer, lognode, checkpointor, etc.) exist in the codebase. However, the **Docker containers, docker-compose configurations, nginx setup, monitoring/alerting infrastructure, and deployment automation** described here are not yet implemented.

## Overview

A full node provides complete Credence network services including trust scoring, event logging, and comprehensive data storage. Ideal for organizations and serious network participants who want to offer full services and earn higher rewards.

## Requirements

### System Requirements
- **CPU**: 8 cores minimum (16+ recommended)
- **RAM**: 16GB minimum (32GB+ recommended)  
- **Storage**: 500GB SSD (1TB+ recommended)
- **Network**: Stable internet, 1Gbps+ preferred, ports 4001-4005/tcp open
- **OS**: Linux (Ubuntu 20.04+ recommended)

### Dependencies
- Docker 20.10+
- Docker Compose 2.0+
- Reliable power supply (UPS recommended)
- Monitoring tools (Prometheus + Grafana optional)

## Quick Start

### 1. Download Configuration
```bash
curl -L https://github.com/ParichayaHQ/credence/releases/latest/download/full-node.tar.gz -o full-node.tar.gz
tar -xzf full-node.tar.gz
cd full-node/
```

### 2. Configure Node
```bash
# Copy example configuration
cp config/full-node.example.yml config/full-node.yml

# Generate node identity
./scripts/generate-identity.sh

# Edit configuration
nano config/full-node.yml
```

### 3. Start Services
```bash
# Start full node stack
docker-compose -f docker-compose.full.yml up -d

# Wait for services to start
sleep 30

# Check status
docker-compose -f docker-compose.full.yml ps
```

### 4. Verify Operation
```bash
# Run health checks
./scripts/health-check.sh

# Verify all services responding
./scripts/service-test.sh
```

## Services Included

### Core Infrastructure
- **P2P Gateway** - Network communication and gossip
- **Full Node** - Complete event storage and retrieval
- **Log Node** - Cryptographic audit trail (Trillian)

### Trust & Scoring  
- **Scorer** - Real-time trust score computation
- **Graph Analytics** - Anti-collusion detection
- **Proof Bundler** - Verification proofs for clients

### Identity Services
- **Issuer** - Verifiable credential issuance (optional)
- **Revocation Registry** - StatusList 2021 management

### Consensus (If Selected)
- **Checkpointor** - Participate in BLS threshold signatures
- **Rules Registry** - Governance proposal execution

## Architecture

```mermaid
graph TB
    Internet[Internet] --> LB[Load Balancer/Nginx]
    
    subgraph "Credence Full Node"
        LB --> PG[P2P Gateway :4001]
        LB --> FN[Full Node :8081]
        LB --> SC[Scorer :8082]
        LB --> LN[Log Node :8083]
        LB --> IS[Issuer :8084]
        LB --> CP[Checkpointor :8085]
        
        PG <--> FN
        FN <--> SC  
        FN <--> LN
        SC <--> LN
        CP <--> FN
        
        FN --> DB[(RocksDB)]
        LN --> TR[(Trillian)]
        SC --> RD[(Redis Cache)]
    end
```

## Configuration

### Full Node Config (`config/full-node.yml`)
```yaml
# Node Identity
identity:
  private_key_path: "/config/keys/node.key"
  public_key_path: "/config/keys/node.pub"
  node_id: "12D3KooW..."

# Network Configuration
network:
  listen_addresses:
    - "/ip4/0.0.0.0/tcp/4001"
  external_address: "/ip4/YOUR_EXTERNAL_IP/tcp/4001"
  bootstrap_peers:
    - "/dns4/seed1.credence.network/tcp/4001/p2p/12D3..."
  max_peers: 500

# Storage Configuration
storage:
  data_dir: "/data"
  retention_policy: "full"      # Keep all historical data
  backup_enabled: true
  backup_interval: "6h"
  compression_enabled: true

# Services Configuration
services:
  p2p_gateway:
    enabled: true
    port: 4001
    api_port: 8080
    
  fullnode:
    enabled: true
    port: 8081
    max_events_per_query: 10000
    cache_size: "8GB"
    
  scorer:
    enabled: true  
    port: 8082
    algorithm_version: "v2.1"
    update_interval: "5m"
    max_graph_depth: 6
    
  lognode:
    enabled: true
    port: 8083
    tree_type: "merkle"
    signing_frequency: "1m"
    
  issuer:
    enabled: false             # Enable if you want to issue VCs
    port: 8084
    
  checkpointor:
    enabled: true              # Auto-enabled if selected as committee member
    port: 8085
    bls_private_key_path: "/config/keys/bls.key"

# Resource Limits
resources:
  max_memory: "12GB"
  max_cpu: "6.0" 
  max_storage: "400GB"
  
# Monitoring
monitoring:
  prometheus_enabled: true
  prometheus_port: 9090
  log_level: "info"
  metrics_retention: "30d"
```

### Docker Compose (`docker-compose.full.yml`)
```yaml
version: '3.8'

services:
  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./config/nginx.conf:/etc/nginx/nginx.conf
      - ./ssl:/etc/ssl/private
    depends_on:
      - p2p-gateway
      - fullnode
      - scorer
    restart: unless-stopped

  p2p-gateway:
    image: credence/p2p-gateway:latest
    ports:
      - "4001:4001"
    volumes:
      - ./config:/config
      - ./data/p2p:/data
      - fullnode-data:/shared-data
    environment:
      - CONFIG_PATH=/config/full-node.yml
      - LOG_LEVEL=info
    restart: unless-stopped

  fullnode:
    image: credence/fullnode:latest
    volumes:
      - ./config:/config
      - ./data/fullnode:/data
      - fullnode-data:/shared-data
    depends_on:
      - p2p-gateway
    environment:
      - CONFIG_PATH=/config/full-node.yml
      - MODE=full
    restart: unless-stopped

  scorer:
    image: credence/scorer:latest
    volumes:
      - ./config:/config
      - ./data/scorer:/data
    depends_on:
      - fullnode
      - redis
    environment:
      - CONFIG_PATH=/config/full-node.yml
      - REDIS_URL=redis://redis:6379
    restart: unless-stopped

  lognode:
    image: credence/lognode:latest
    volumes:
      - ./config:/config  
      - ./data/lognode:/data
    depends_on:
      - fullnode
      - postgres
    environment:
      - CONFIG_PATH=/config/full-node.yml
      - DATABASE_URL=postgres://credence:password@postgres/trillian
    restart: unless-stopped

  checkpointor:
    image: credence/checkpointor:latest
    volumes:
      - ./config:/config
      - ./data/checkpointor:/data
    depends_on:
      - fullnode
    environment:
      - CONFIG_PATH=/config/full-node.yml
    restart: unless-stopped

  # Supporting Services
  redis:
    image: redis:7-alpine
    volumes:
      - redis-data:/data
    command: redis-server --appendonly yes
    restart: unless-stopped

  postgres:
    image: postgres:15-alpine
    volumes:
      - postgres-data:/var/lib/postgresql/data
    environment:
      - POSTGRES_DB=trillian
      - POSTGRES_USER=credence
      - POSTGRES_PASSWORD=password
    restart: unless-stopped

  # Monitoring
  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./config/prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus-data:/prometheus
    restart: unless-stopped

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    volumes:
      - grafana-data:/var/lib/grafana
      - ./config/grafana:/etc/grafana/provisioning
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    restart: unless-stopped

volumes:
  fullnode-data:
  redis-data:
  postgres-data:
  prometheus-data:
  grafana-data:
```

## SSL/TLS Configuration

### Generate Certificates
```bash
# Self-signed for development
./scripts/generate-ssl.sh

# Let's Encrypt for production
certbot certonly --standalone -d your-node.example.com
```

### Nginx Configuration (`config/nginx.conf`)
```nginx
events {
    worker_connections 1024;
}

http {
    upstream credence_api {
        server fullnode:8081;
        server scorer:8082;
    }
    
    server {
        listen 443 ssl http2;
        server_name your-node.example.com;
        
        ssl_certificate /etc/ssl/private/cert.pem;
        ssl_certificate_key /etc/ssl/private/key.pem;
        
        location /api/v1/ {
            proxy_pass http://credence_api;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
        }
        
        location /metrics {
            proxy_pass http://prometheus:9090;
            auth_basic "Metrics";
            auth_basic_user_file /etc/nginx/.htpasswd;
        }
    }
}
```

## Management

### Service Operations
```bash
# Start all services
docker-compose -f docker-compose.full.yml up -d

# Stop gracefully
docker-compose -f docker-compose.full.yml down

# Update services
docker-compose -f docker-compose.full.yml pull
docker-compose -f docker-compose.full.yml up -d

# Scale specific services
docker-compose -f docker-compose.full.yml up -d --scale scorer=2
```

### Backup & Recovery
```bash
# Create backup
./scripts/backup.sh

# Restore from backup
./scripts/restore.sh /path/to/backup.tar.gz

# Automated backup schedule (add to crontab)
0 2 * * * /path/to/full-node/scripts/backup.sh
```

### Log Management
```bash
# View service logs
docker-compose -f docker-compose.full.yml logs -f scorer

# Export logs for analysis
docker-compose -f docker-compose.full.yml logs --since 24h > node-logs.txt

# Log rotation (configure in docker-compose)
logging:
  driver: "json-file"
  options:
    max-size: "100m"
    max-file: "10"
```

## Monitoring & Alerts

### Prometheus Metrics
Key metrics to monitor:
- **p2p_peers_connected** - Network connectivity
- **fullnode_events_stored** - Data ingestion rate
- **scorer_computations_per_second** - Processing capacity
- **system_memory_usage** - Resource utilization

### Grafana Dashboards
Pre-configured dashboards included:
- **Node Overview** - High-level health metrics
- **Network Stats** - P2P and sync status
- **Performance** - CPU, memory, disk usage
- **Business Metrics** - Trust scores, events processed

### Alerting Rules
```yaml
# config/alert-rules.yml
groups:
  - name: credence-node
    rules:
      - alert: NodeOffline  
        expr: up == 0
        for: 2m
        
      - alert: LowDiskSpace
        expr: disk_free_bytes / disk_total_bytes < 0.1
        for: 5m
        
      - alert: HighMemoryUsage
        expr: memory_usage_ratio > 0.9
        for: 10m
```

## Economics

### Rewards Structure
- **Base reward**: 100 tokens/month for 99% uptime
- **Performance bonus**: +50% for top 25% of nodes
- **Service multipliers**:
  - P2P Gateway: 1.0x
  - Full Node: 2.0x  
  - Scorer: 1.5x
  - Log Node: 1.2x
  - Checkpointor: 3.0x (when active)

### Operating Costs
- **Hardware**: $200-500/month (VPS/dedicated)
- **Bandwidth**: $50-200/month
- **Maintenance**: ~4-8 hours/month

### ROI Calculation
```bash
# Use the reward calculator
./scripts/calculate-rewards.sh --uptime 99.5 --services all --region us-east
```

## Security

### Hardening Checklist
- [ ] OS updates and security patches current
- [ ] SSH key-based authentication only
- [ ] Firewall configured (only necessary ports open)
- [ ] SSL/TLS certificates valid and renewed
- [ ] Regular backup testing
- [ ] Monitoring and alerting operational
- [ ] Log aggregation and retention policy
- [ ] Incident response plan documented

### Key Management
```bash
# Secure key generation
./scripts/generate-keys.sh --secure

# Key rotation (quarterly recommended)
./scripts/rotate-keys.sh --backup-old

# Hardware Security Module (optional)
./scripts/setup-hsm.sh --provider yubihsm2
```

## Troubleshooting

### Common Issues

**Services failing to start:**
```bash
# Check resource constraints
docker system df
free -h
df -h

# Review service logs
docker-compose -f docker-compose.full.yml logs --tail 100 scorer
```

**Poor performance:**
```bash
# Check if storage is bottleneck
iotop -a

# Monitor network utilization  
iftop

# Profile specific service
docker stats $(docker-compose -f docker-compose.full.yml ps -q scorer)
```

**Network connectivity issues:**
```bash
# Test P2P connectivity
./scripts/test-p2p.sh

# Check bootstrap peer connectivity
./scripts/test-bootstrap.sh

# Verify external port accessibility
./scripts/test-external-ports.sh
```

## Scaling & Optimization

### Horizontal Scaling
```bash
# Run multiple scorer instances
docker-compose -f docker-compose.full.yml up -d --scale scorer=3

# Load balance with nginx
# Add upstream servers to nginx.conf
```

### Performance Tuning
```bash
# Optimize for SSD storage
echo 'deadline' > /sys/block/sda/queue/scheduler

# Tune network buffer sizes
sysctl -w net.core.rmem_max=134217728
sysctl -w net.core.wmem_max=134217728

# Configure Docker resource limits
# Edit docker-compose.yml resource constraints
```

Running a full node makes you a key infrastructure provider for the Credence network!