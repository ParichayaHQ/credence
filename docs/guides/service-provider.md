# Service Provider Deployment Guide

> **⚠️ IMPLEMENTATION STATUS**  
> The issuer service exists in the codebase. However, the **Docker containers, configuration templates, credential schemas, web portals, deployment packages, and integration examples** described here are not yet implemented. The core issuer functionality exists but requires manual setup.

## Overview

Service providers run specialized services for specific use cases like credential issuance, identity verification, or custom trust scoring. Perfect for organizations that want to offer identity services to their community.

## Use Cases

### Educational Institutions
- **Diploma/Certificate Issuance** - Issue verifiable academic credentials
- **Student Identity** - Manage student DIDs and verification
- **Academic Transcripts** - Tamper-proof academic records

### Employers & HR
- **Employment Verification** - Issue work history credentials
- **Skills Certification** - Validate professional competencies  
- **Background Checks** - Trusted verification services

### Government Agencies
- **Identity Documents** - Driver's licenses, passports, permits
- **Licenses & Permits** - Professional licenses, business permits
- **Public Records** - Birth certificates, marriage licenses

### Healthcare Providers
- **Medical Credentials** - Doctor/nurse licensing
- **Vaccination Records** - COVID-19 and other immunization proof
- **Health Insurance** - Coverage verification

## Architecture Options

### Issuer-Only Service
Minimal setup for credential issuance:
```
[Your Application] → [Issuer Service] → [P2P Network]
```

### Full Identity Provider
Complete identity management platform:
```
[Web Portal] → [Issuer + Verifier + Registry] → [P2P Network]
```

### Specialized Verifier
Custom verification logic for specific industries:
```
[Your App] → [Custom Verifier] → [Credence Network] → [Response]
```

## Quick Start - Issuer Service

### 1. Download Issuer Template
```bash
curl -L https://github.com/ParichayaHQ/credence/releases/latest/download/issuer-service.tar.gz -o issuer-service.tar.gz
tar -xzf issuer-service.tar.gz
cd issuer-service/
```

### 2. Configure Your Organization
```bash
# Copy configuration template
cp config/issuer.example.yml config/issuer.yml

# Generate organizational keys
./scripts/generate-org-keys.sh --org "University of Example"

# Edit configuration
nano config/issuer.yml
```

### 3. Customize Credential Schemas
```bash
# Edit credential templates
ls config/schemas/
# - diploma.json
# - certificate.json  
# - employment.json
# - custom.json

# Add your organization's schemas
nano config/schemas/your-credential-type.json
```

### 4. Deploy Service
```bash
# Start issuer service
docker-compose -f docker-compose.issuer.yml up -d

# Test issuance
curl -X POST http://localhost:8084/issue \
  -H "Content-Type: application/json" \
  -d @test-data/sample-credential.json
```

## Configuration

### Issuer Configuration (`config/issuer.yml`)
```yaml
# Organization Identity
organization:
  name: "University of Example"
  did: "did:web:university.example.com"
  domain: "university.example.com"
  logo_url: "https://university.example.com/logo.png"
  
# Signing Configuration
signing:
  method: "Ed25519"
  key_path: "/config/keys/issuer.key"
  key_id: "key-1"
  
# Credential Templates
credentials:
  - type: "UniversityDiploma"
    schema_path: "/config/schemas/diploma.json"
    template_path: "/config/templates/diploma.json"
    validity_period: "P4Y"  # 4 years
    
  - type: "CourseCertificate" 
    schema_path: "/config/schemas/certificate.json"
    template_path: "/config/templates/certificate.json"
    validity_period: "P1Y"  # 1 year
    
# Revocation Configuration
revocation:
  enabled: true
  status_list_url: "https://university.example.com/status"
  update_interval: "1h"
  
# API Configuration
api:
  listen: "0.0.0.0:8084"
  tls_enabled: true
  cors_origins: 
    - "https://university.example.com"
    - "https://student-portal.example.com"
  rate_limit: "100/hour/ip"
  
# Authentication
auth:
  method: "api_key"  # or "oauth2", "saml"
  api_keys:
    - key: "your-api-key-here"
      scope: ["issue", "revoke"]
      description: "Student Records System"
      
# Integration
integration:
  webhook_url: "https://university.example.com/webhooks/credence"
  database_url: "postgres://user:pass@localhost/university_records"
  
# Network
network:
  p2p_endpoint: "http://p2p-gateway:8080"
  bootstrap: true
  announce_service: true
```

### Credential Schema Example (`config/schemas/diploma.json`)
```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "title": "University Diploma",
  "properties": {
    "credentialSubject": {
      "type": "object",
      "properties": {
        "id": {
          "type": "string",
          "format": "uri",
          "description": "Student DID"
        },
        "name": {
          "type": "string",
          "description": "Student full name"
        },
        "degree": {
          "type": "string",
          "enum": ["Bachelor of Science", "Bachelor of Arts", "Master of Science", "Doctor of Philosophy"]
        },
        "major": {
          "type": "string",
          "description": "Field of study"
        },
        "graduationDate": {
          "type": "string",
          "format": "date"
        },
        "gpa": {
          "type": "number",
          "minimum": 0.0,
          "maximum": 4.0
        },
        "honors": {
          "type": "string",
          "enum": ["summa cum laude", "magna cum laude", "cum laude"]
        }
      },
      "required": ["id", "name", "degree", "major", "graduationDate"]
    }
  }
}
```

### Credential Template (`config/templates/diploma.json`)
```json
{
  "@context": [
    "https://www.w3.org/2018/credentials/v1",
    "https://credence.network/contexts/education/v1"
  ],
  "type": ["VerifiableCredential", "UniversityDiploma"],
  "issuer": {
    "id": "did:web:university.example.com",
    "name": "University of Example",
    "image": "https://university.example.com/logo.png"
  },
  "issuanceDate": "{{NOW}}",
  "expirationDate": "{{ISSUE_DATE + 4 YEARS}}",
  "credentialStatus": {
    "type": "StatusList2021Entry",
    "statusListIndex": "{{AUTO_INCREMENT}}",
    "statusListCredential": "https://university.example.com/status/1"
  },
  "credentialSubject": {
    "id": "{{STUDENT_DID}}",
    "name": "{{STUDENT_NAME}}",
    "degree": "{{DEGREE_TYPE}}",
    "major": "{{MAJOR_FIELD}}",
    "graduationDate": "{{GRADUATION_DATE}}",
    "institution": {
      "name": "University of Example",
      "id": "did:web:university.example.com"
    }
  }
}
```

## Docker Deployment

### Issuer Service (`docker-compose.issuer.yml`)
```yaml
version: '3.8'

services:
  issuer:
    image: credence/issuer:latest
    ports:
      - "8084:8084"
      - "443:443"
    volumes:
      - ./config:/config
      - ./data:/data
      - ./ssl:/ssl
    environment:
      - CONFIG_PATH=/config/issuer.yml
      - LOG_LEVEL=info
      - TLS_CERT_PATH=/ssl/cert.pem
      - TLS_KEY_PATH=/ssl/key.pem
    restart: unless-stopped
    
  # Optional: Connect to existing Credence network
  p2p-gateway:
    image: credence/p2p-gateway:latest
    ports:
      - "4001:4001"
    volumes:
      - ./config:/config
      - ./data/p2p:/data
    environment:
      - CONFIG_PATH=/config/issuer.yml
      - MODE=client
    restart: unless-stopped
    
  # Database for credential tracking
  postgres:
    image: postgres:15-alpine
    volumes:
      - postgres-data:/var/lib/postgresql/data
    environment:
      - POSTGRES_DB=issuer
      - POSTGRES_USER=credence
      - POSTGRES_PASSWORD=secure-password
    restart: unless-stopped
    
  # Redis for caching and sessions
  redis:
    image: redis:7-alpine
    volumes:
      - redis-data:/data
    restart: unless-stopped

volumes:
  postgres-data:
  redis-data:
```

## Integration Examples

### Student Information System
```python
import requests
import json

# Issue diploma when student graduates
def issue_diploma(student_data):
    credential = {
        "credentialSubject": {
            "id": student_data["did"],
            "name": student_data["full_name"],
            "degree": "Bachelor of Science",
            "major": student_data["major"],
            "graduationDate": student_data["graduation_date"],
            "gpa": student_data["gpa"]
        }
    }
    
    response = requests.post(
        "https://your-issuer.university.com/api/v1/issue",
        headers={
            "Authorization": "Bearer your-api-key",
            "Content-Type": "application/json"
        },
        json=credential
    )
    
    return response.json()
```

### HR System Integration
```javascript
// Issue employment verification
async function issueEmploymentCredential(employeeId, position, startDate) {
  const credential = {
    credentialSubject: {
      id: `did:key:${employeeId}`,
      name: employee.fullName,
      position: position,
      employer: "Acme Corporation",
      startDate: startDate,
      department: employee.department,
      employmentType: "full-time"
    }
  };
  
  const response = await fetch('/api/credentials/issue', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-API-Key': process.env.CREDENCE_API_KEY
    },
    body: JSON.stringify(credential)
  });
  
  return response.json();
}
```

### Government Portal
```bash
# Bulk issuance for licenses
curl -X POST https://licenses.gov.example/api/bulk-issue \
  -H "Authorization: Bearer gov-api-token" \
  -F "file=@driver-licenses-batch.csv" \
  -F "credential_type=DriversLicense"
```

## Web Portal Setup

### Admin Dashboard
```bash
# Deploy web interface for credential management
docker-compose -f docker-compose.portal.yml up -d

# Access at https://your-domain.com/admin
# Features:
# - Issue credentials manually
# - View issuance history
# - Manage revocations
# - Analytics dashboard
```

### Self-Service Portal
```bash
# Allow users to request credentials
# Configure in config/portal.yml:

portal:
  self_service:
    enabled: true
    workflows:
      - type: "diploma_request"
        approval_required: true
        auto_approve_verified_students: true
      - type: "transcript_request"  
        approval_required: false
        fee: "$5.00"
```

## Verification Services

### Custom Verifier
```yaml
# config/verifier.yml
verifier:
  policies:
    - name: "employment_verification"
      description: "Verify current employment status"
      required_credentials:
        - type: "EmploymentCredential"
          issuer_trust_score: "> 0.8"
          max_age: "P6M"  # 6 months
      response_format: "minimal"
      
    - name: "education_verification"
      description: "Verify educational qualifications"  
      required_credentials:
        - type: "UniversityDiploma"
          accredited_institutions_only: true
        - type: "Transcript"
          min_gpa: 3.0
```

## Monitoring & Analytics

### Issuance Metrics
```bash
# View credential issuance stats
curl https://your-issuer.com/api/metrics | jq .

{
  "credentials_issued_today": 47,
  "total_credentials_issued": 12845,
  "active_credentials": 12443,
  "revoked_credentials": 402,
  "most_issued_type": "UniversityDiploma",
  "average_issuance_time": "0.8s"
}
```

### Trust Score Impact
```bash
# Check how your issuances affect trust scores
./scripts/trust-impact-report.sh --period 30d

# Output:
# Average trust score increase: +0.15
# Credentials verified: 3,247
# Verification success rate: 99.2%
# Network trust contribution: High
```

## Business Models

### Subscription Based
- **Monthly fee per credential type**
- **Usage tiers**: Basic, Professional, Enterprise
- **Volume discounts** for large organizations

### Pay-per-Credential
- **$0.50-2.00 per credential issued**
- **Bulk pricing** for 1000+ credentials
- **Premium features** (faster issuance, custom branding)

### White Label
- **License the service** to other organizations
- **Revenue sharing** model
- **Custom branding** and domain

## Security & Compliance

### Key Management
```bash
# Rotate signing keys quarterly
./scripts/rotate-issuer-keys.sh --backup-old

# Use Hardware Security Module (recommended for production)
./scripts/setup-hsm.sh --provider aws-cloudhsm

# Multi-signature for high-value credentials
./scripts/enable-multisig.sh --threshold 2 --signers 3
```

### Audit & Compliance
```bash
# Generate compliance report
./scripts/compliance-report.sh --standard SOC2 --period Q1-2024

# Export credential history for audit
./scripts/export-audit-log.sh --format json --encrypted
```

### Privacy Protection
- **Minimal data collection** - Only what's needed for issuance
- **Data retention policies** - Automatic deletion after set period
- **Zero-knowledge proofs** - Enable selective disclosure
- **GDPR compliance** - Right to be forgotten implementation

## Scaling & Performance

### High Availability
```yaml
# Load balanced issuer setup
services:
  issuer:
    deploy:
      replicas: 3
      update_config:
        parallelism: 1
        delay: 10s
      restart_policy:
        condition: on-failure
```

### Performance Optimization
```bash
# Batch credential issuance
curl -X POST /api/v1/batch-issue \
  -H "Content-Type: application/json" \
  -d @bulk-credentials.json

# Async processing for large batches
./scripts/queue-bulk-issuance.sh --file graduates-2024.csv --async
```

Service providers enable organizations to seamlessly integrate verifiable credentials into their existing workflows while contributing to the decentralized trust network!