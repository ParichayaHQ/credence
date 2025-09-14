# Credence Deployment Commands

Quick reference for deploying different Credence configurations.

## 🔧 End User Wallet (Desktop)

### Development Mode
```bash
# Clone and setup desktop wallet
git clone https://github.com/ParichayaHQ/credence.git
cd credence/desktop

# Install dependencies and run
yarn install
yarn dev

# Open Electron app at http://localhost (auto-opens)
```

### Production Release (Future)
```bash
# Download installers (when available)
curl -L https://github.com/ParichayaHQ/credence/releases/latest/download/Credence-mac.dmg -o Credence.dmg
open Credence.dmg

# Or Windows
curl -L https://github.com/ParichayaHQ/credence/releases/latest/download/Credence-win.msi -o Credence.msi

# Or Linux
curl -L https://github.com/ParichayaHQ/credence/releases/latest/download/Credence-linux.AppImage -o Credence.AppImage
chmod +x Credence.AppImage && ./Credence.AppImage
```

---

## 🏗️ Light Node Deployment (30 min setup)

### 1. Download Configuration
```bash
# Create deployment directory
mkdir credence-light-node
cd credence-light-node

# Download light node package (future)
curl -L https://github.com/ParichayaHQ/credence/releases/latest/download/light-node.tar.gz -o light-node.tar.gz
tar -xzf light-node.tar.gz

# Or clone repo for now
git clone https://github.com/ParichayaHQ/credence.git
cd credence
```

### 2. Configure Node
```bash
# Copy configuration template
mkdir -p config
cp config-templates/light-node.example.yml config/light-node.yml

# Edit configuration
nano config/light-node.yml
# - Set YOUR_EXTERNAL_IP
# - Set node_operator_address for rewards
# - Adjust resource limits if needed
```

### 3. Start Services
```bash
# Start light node stack
docker-compose -f docker-compose.light.yml up -d

# Check status
docker-compose -f docker-compose.light.yml ps

# View logs
docker-compose -f docker-compose.light.yml logs -f
```

### 4. Verify Operation
```bash
# Check P2P connectivity
curl http://localhost:8080/health

# Check peer connections
curl http://localhost:8080/peers | jq .

# Check data sync status  
curl http://localhost:8081/sync/status | jq .

# Monitor resource usage
docker stats
```

### Management Commands
```bash
# Stop services
docker-compose -f docker-compose.light.yml down

# Update to latest
docker-compose -f docker-compose.light.yml pull
docker-compose -f docker-compose.light.yml up -d

# View service logs
docker-compose -f docker-compose.light.yml logs -f p2p-gateway
docker-compose -f docker-compose.light.yml logs -f fullnode

# Check storage usage
du -sh data/
```

---

## 🏢 Full Node Deployment (2-4 hour setup)

### 1. Download Configuration
```bash
# Create deployment directory
mkdir credence-full-node
cd credence-full-node

# Download full node package (future)
curl -L https://github.com/ParichayaHQ/credence/releases/latest/download/full-node.tar.gz -o full-node.tar.gz
tar -xzf full-node.tar.gz

# Or clone repo for now
git clone https://github.com/ParichayaHQ/credence.git
cd credence
```

### 2. Configure Node
```bash
# Copy configuration template
mkdir -p config ssl
cp config-templates/full-node.example.yml config/full-node.yml

# Generate node identity
./scripts/generate-identity.sh

# Generate SSL certificates
./scripts/generate-ssl.sh your-domain.com

# Edit configuration
nano config/full-node.yml
# - Set YOUR_EXTERNAL_IP  
# - Set node_operator_address
# - Configure SSL paths
# - Set PostgreSQL password
```

### 3. Environment Setup
```bash
# Create environment file
cat > .env << EOF
POSTGRES_PASSWORD=your-secure-password
GRAFANA_PASSWORD=admin-password
EOF

# Set file permissions
chmod 600 .env
chmod 600 config/full-node.yml
```

### 4. Start Services
```bash
# Start full node stack with monitoring
docker-compose -f docker-compose.full.yml --profile with-monitoring up -d

# Or start without monitoring
docker-compose -f docker-compose.full.yml up -d

# Check all services are healthy
docker-compose -f docker-compose.full.yml ps
```

### 5. Verify Operation
```bash
# Run comprehensive health check
./scripts/health-check.sh

# Check individual services
curl https://your-domain.com/api/v1/health
curl http://localhost:9090  # Prometheus
curl http://localhost:3000  # Grafana (admin/admin-password)

# Monitor metrics
curl http://localhost:9090/api/v1/query?query=up
```

### Management Commands
```bash
# View all logs
docker-compose -f docker-compose.full.yml logs -f

# Scale specific service
docker-compose -f docker-compose.full.yml up -d --scale scorer=2

# Backup data
./scripts/backup.sh

# Update services
docker-compose -f docker-compose.full.yml pull
docker-compose -f docker-compose.full.yml up -d

# Emergency shutdown
docker-compose -f docker-compose.full.yml down
```

---

## 🏢 Service Provider (Issuer) Deployment

### 1. Setup Issuer Service
```bash
# Clone repository
git clone https://github.com/ParichayaHQ/credence.git
cd credence

# Download issuer template (future)
curl -L https://github.com/ParichayaHQ/credence/releases/latest/download/issuer-service.tar.gz -o issuer-service.tar.gz
tar -xzf issuer-service.tar.gz
```

### 2. Configure Organization
```bash
# Copy issuer configuration
mkdir -p config/schemas config/templates
cp config-templates/issuer.example.yml config/issuer.yml

# Generate organizational keys
./scripts/generate-org-keys.sh --org "University of Example"

# Edit configuration
nano config/issuer.yml
# - Set organization details
# - Configure credential schemas
# - Set API authentication
```

### 3. Deploy Issuer
```bash
# Start with full node (includes P2P connectivity)
docker-compose -f docker-compose.full.yml --profile with-issuer up -d

# Or standalone issuer (requires external P2P)
docker-compose -f docker-compose.issuer.yml up -d
```

### 4. Test Credential Issuance
```bash
# Test API endpoint
curl -X POST https://your-domain.com/api/v1/issue \
  -H "Authorization: Bearer your-api-key" \
  -H "Content-Type: application/json" \
  -d @test-data/sample-credential.json

# Check credential status
curl https://your-domain.com/status/1
```

---

## 👩‍💻 Development Environment

### 1. Local Development Stack
```bash
# Clone repository
git clone https://github.com/ParichayaHQ/credence.git
cd credence

# Start complete development environment
docker-compose -f docker-compose.dev.yml up -d

# Wait for services to be ready
docker-compose -f docker-compose.dev.yml ps
```

### 2. Access Development Services
```bash
# Service endpoints
curl http://localhost:8080  # P2P Gateway
curl http://localhost:8081  # Full Node
curl http://localhost:8082  # Scorer
curl http://localhost:8083  # Log Node  
curl http://localhost:8084  # Wallet Service
curl http://localhost:8085  # Checkpointor

# Monitoring
open http://localhost:3000   # Grafana (admin/admin)
open http://localhost:9090   # Prometheus
```

### 3. Development Workflow
```bash
# View live logs
docker-compose -f docker-compose.dev.yml logs -f

# Rebuild specific service
docker-compose -f docker-compose.dev.yml build p2p-gateway
docker-compose -f docker-compose.dev.yml up -d p2p-gateway

# Reset development data
docker-compose -f docker-compose.dev.yml down -v
docker-compose -f docker-compose.dev.yml up -d
```

---

## 🛠️ Maintenance Commands

### Health Monitoring
```bash
# Check all service health
for port in 8080 8081 8082 8083 8084 8085; do
  echo "Checking localhost:$port/health"
  curl -s http://localhost:$port/health | jq .
done

# Check Docker container stats
docker stats --no-stream

# Check disk usage
df -h
du -sh data/
```

### Log Management
```bash
# View recent logs
docker-compose logs --tail=100 -f

# Export logs for analysis
docker-compose logs --since=24h > credence-logs.txt

# Clear old logs (careful!)
docker system prune --volumes
```

### Backup & Recovery
```bash
# Create backup
./scripts/backup.sh full

# List backups
ls -la backups/

# Restore from backup
./scripts/restore.sh backups/credence-backup-2024-01-15.tar.gz
```

### Updates
```bash
# Pull latest images
docker-compose pull

# Update with zero downtime (full node)
./scripts/rolling-update.sh

# Update with downtime
docker-compose down
docker-compose pull  
docker-compose up -d
```

---

## 🔧 Troubleshooting

### Common Issues
```bash
# Services won't start
docker-compose logs SERVICE_NAME

# Port conflicts
sudo netstat -tulpn | grep :4001
sudo lsof -i :4001

# Disk space issues
docker system df
docker system prune

# Permission issues
sudo chown -R $USER:$USER data/
chmod -R 755 data/

# Network connectivity
nc -zv your-external-ip 4001
./scripts/test-p2p-connectivity.sh
```

### Performance Issues
```bash
# Monitor resource usage
docker stats --no-stream
htop
iotop -a

# Check service health
./scripts/performance-report.sh

# Restart problematic service
docker-compose restart scorer
```

## 📱 Quick Setup Summary

| **Use Case** | **Time** | **Command** |
|-------------|----------|-------------|
| **Desktop Wallet** | 5 min | `cd desktop && yarn install && yarn dev` |
| **Light Node** | 30 min | `docker-compose -f docker-compose.light.yml up -d` |
| **Full Node** | 2-4 hours | `docker-compose -f docker-compose.full.yml up -d` |
| **Development** | 10 min | `docker-compose -f docker-compose.dev.yml up -d` |
| **Service Provider** | 1-2 hours | `docker-compose -f docker-compose.full.yml --profile with-issuer up -d` |

Each setup includes proper configuration templates, health checks, and monitoring! 🚀