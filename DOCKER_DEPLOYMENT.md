# Nexus Docker Deployment Guide

Complete guide for running Nexus locally (development/testing) and in production.

---

## Table of Contents
- [Local Development](#local-development)
- [Production Deployment](#production-deployment)
- [Agent Deployment (Customer)](#agent-deployment-customer)
- [Troubleshooting](#troubleshooting)

---

## Local Development

### Prerequisites
- Docker & Docker Compose installed
- Git with submodule support

### Step 1: Clone Repository

```bash
git clone --recurse-submodules https://github.com/YOUR_ORG/nexus-architecture.git
cd nexus-architecture
```

### Step 2: Start MySQL & Redis

```bash
# Create network
docker network create nexus-infra_default

# Start MySQL
docker run -d \
  --name nexus_mysql_local \
  --network nexus-infra_default \
  -e MYSQL_ROOT_PASSWORD=root \
  -e MYSQL_DATABASE=nexus_db \
  -e MYSQL_USER=demo_user \
  -e MYSQL_PASSWORD=demo_password \
  -p 3309:3306 \
  mysql:8.0

# Start Redis
docker run -d \
  --name nexus_redis_local \
  --network nexus-infra_default \
  -p 6381:6379 \
  redis:alpine
```

### Step 3: Create Environment Files

**nexus-core/.env.docker:**
```env
NEXUS_DB_HOST=nexus_mysql_local
NEXUS_DB_PORT=3306
NEXUS_DB_USER=demo_user
NEXUS_DB_PASSWORD=demo_password
NEXUS_DB_NAME=nexus_db

NEXUS_REDIS_HOST=nexus_redis_local
NEXUS_REDIS_PORT=6379

NEXUS_JWT_SECRET=your-local-jwt-secret-here
NEXUS_ENCRYPTION_KEY=your-local-32-byte-key-base64

GIN_MODE=debug
PORT=8080
```

### Step 4: Start Nexus Services

```bash
docker-compose -f docker-compose.local.yml up --build
```

### Step 5: Access

| Service | URL |
|---------|-----|
| Nexus UI | http://localhost:3000 |
| Nexus API | http://localhost:8080 |
| Nginx | http://localhost |

### Step 6: Stop

```bash
docker-compose -f docker-compose.local.yml down

# To also remove data:
docker stop nexus_mysql_local nexus_redis_local
docker rm nexus_mysql_local nexus_redis_local
```

---

## Production Deployment

### Prerequisites
- Ubuntu 22.04+ server
- Docker & Docker Compose
- MySQL server (can be separate)
- Redis server (can be separate)
- Domain name (optional)

### Step 1: Server Setup

```bash
# Install Docker
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER

# Install Docker Compose
sudo apt install docker-compose-plugin -y
```

### Step 2: Clone Repository

```bash
cd /opt
git clone --recurse-submodules https://github.com/YOUR_ORG/nexus-architecture.git
cd nexus-architecture
```

### Step 3: Configure Environment

**nexus-core/.env:**
```env
# Database (your MySQL server)
NEXUS_DB_HOST=YOUR_MYSQL_HOST
NEXUS_DB_PORT=3306
NEXUS_DB_USER=nexus_admin
NEXUS_DB_PASSWORD=SECURE_PASSWORD
NEXUS_DB_NAME=nexus_db

# Redis
NEXUS_REDIS_HOST=YOUR_REDIS_HOST
NEXUS_REDIS_PORT=6379

# Security (generate: openssl rand -base64 32)
NEXUS_JWT_SECRET=GENERATED_JWT_SECRET
NEXUS_ENCRYPTION_KEY=GENERATED_ENCRYPTION_KEY

GIN_MODE=release
PORT=8080
```

**nexus-core/config.yml:**
```yaml
server:
  port: "8080"

redis:
  addr: "YOUR_REDIS_HOST:6379"
  streamname: "nexus_stream"
  consumergroup: "nexus_group"

database:
  nexus_db_dsn: "nexus_admin:PASSWORD@tcp(YOUR_MYSQL_HOST:3306)/nexus_db?parseTime=true"

mqtt:
  enabled: true
  reconnect_interval: "5s"
  max_reconnect_attempts: 10
```

### Step 4: Update UI Dockerfile

Edit `nexus_ui/Dockerfile`, update line ~38:
```dockerfile
ENV NEXT_PUBLIC_API_URL=http://YOUR_SERVER_IP:8080
```

### Step 5: Grant MySQL Access

On MySQL server:
```sql
CREATE USER 'nexus_admin'@'%' IDENTIFIED BY 'SECURE_PASSWORD';
GRANT ALL PRIVILEGES ON nexus_db.* TO 'nexus_admin'@'%';
FLUSH PRIVILEGES;
```

### Step 6: Run Database Migrations

Connect to MySQL and run migration SQL files from `nexus_ui/prisma/migrations/`.

### Step 7: Build and Deploy

```bash
docker compose build --no-cache
docker compose up -d
```

### Step 8: Verify

```bash
# Check status
docker compose ps

# Check logs
docker compose logs -f

# Health check
curl http://localhost:8080/health
```

### Step 9: Setup Firewall

```bash
sudo ufw allow 80
sudo ufw allow 443
sudo ufw allow 8080
```

---

## Agent Deployment (Customer)

### Download Agent

```bash
# Linux AMD64
wget https://github.com/YOUR_ORG/nexus-architecture/releases/latest/download/nexus-agent-linux-amd64
chmod +x nexus-agent-linux-amd64
sudo mv nexus-agent-linux-amd64 /usr/local/bin/nexus-agent
```

### Create Configuration

```bash
sudo mkdir -p /etc/nexus-agent
sudo nano /etc/nexus-agent/config.yml
```

```yaml
agent:
  port: 9000
  bind: "0.0.0.0"

nexus:
  server_url: "http://YOUR_NEXUS_SERVER:8080"
  agent_token: "agt_TOKEN_FROM_NEXUS_UI"
  sync_interval: 60s
  timeout: 30s

buffer:
  enabled: true
  max_size: 10000
  db_path: "/var/lib/nexus/queue.db"
```

### Run as Service

```bash
sudo nano /etc/systemd/system/nexus-agent.service
```

```ini
[Unit]
Description=Nexus Agent
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/nexus-agent -config /etc/nexus-agent/config.yml
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable nexus-agent
sudo systemctl start nexus-agent
```

### Run Agent with Docker

```bash
docker run -d \
  --name nexus-agent \
  -p 9000:9000 \
  -v /path/to/config.yml:/etc/nexus-agent/config.yml \
  nexus-agent
```

### Test Agent

```bash
curl http://localhost:9000/health

curl -X POST http://localhost:9000/send \
  -H "Content-Type: application/json" \
  -d '{"app_key": "app_xxx", "data": {"test": "hello"}}'
```

---

## Troubleshooting

| Issue | Solution |
|-------|----------|
| Container won't start | `docker compose logs SERVICE_NAME` |
| Database connection refused | Check MySQL host/port/firewall |
| Access denied MySQL | Grant user from '%' host |
| Redis connection refused | Check Redis host/port |
| UI login not hitting API | Rebuild UI with correct `NEXT_PUBLIC_API_URL` |
| Agent health fails | Check config.yml `bind: "0.0.0.0"` |
| Decryption error | Ensure `NEXUS_ENCRYPTION_KEY` is consistent |

### Common Commands

```bash
# Restart service
docker compose restart nexus-api

# Rebuild single service
docker compose build nexus-api --no-cache
docker compose up -d nexus-api

# View logs
docker compose logs -f nexus-worker

# Enter container
docker exec -it nexus-api sh

# Check network
docker network ls
docker network inspect nexus-architecture_nexus-network
```

---

## Security Notes

> ⚠️ **Important Security Reminders**

1. **Never commit `.env` files** - They contain secrets
2. **Generate unique keys for production** - Use `openssl rand -base64 32`
3. **Never change `NEXUS_ENCRYPTION_KEY`** after data is created
4. **Use HTTPS in production** - Configure SSL with nginx
5. **Restrict database access** - Only allow from application server IP
