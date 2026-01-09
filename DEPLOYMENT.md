# Nexus Deployment Guide

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                    APPLICATION SERVER                            │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐              │
│  │  Nginx :80  │  │  API :8080  │  │  UI :3000   │              │
│  │   (Proxy)   │──│  (Go API)   │  │ (Next.js)   │              │
│  └─────────────┘  └─────────────┘  └─────────────┘              │
│                   ┌─────────────┐                                │
│                   │   Worker    │                                │
│                   │ (Go Worker) │                                │
│                   └─────────────┘                                │
└───────────────────────────┬─────────────────────────────────────┘
                            │ Private Network
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                     STORAGE SERVER                               │
│         ┌─────────────┐        ┌─────────────┐                  │
│         │ MySQL :3306 │        │ Redis :6379 │                  │
│         └─────────────┘        └─────────────┘                  │
└─────────────────────────────────────────────────────────────────┘
```

## Prerequisites

### Application Server
- Ubuntu 20.04+ 
- Docker & Docker Compose installed
- At least 2GB RAM, 2 CPU cores
- Network access to Storage Server

### Storage Server
- MySQL 8.0+ running and accessible
- Redis 6.0+ running and accessible
- Firewall allows connections from Application Server

## Step 1: Prepare Storage Server

### 1.1 MySQL Configuration

```sql
-- Create database and user for Nexus
CREATE DATABASE nexus CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'nexus'@'%' IDENTIFIED BY 'YOUR_SECURE_PASSWORD';
GRANT ALL PRIVILEGES ON nexus.* TO 'nexus'@'%';
FLUSH PRIVILEGES;
```

Make sure MySQL allows remote connections:
```bash
# Edit /etc/mysql/mysql.conf.d/mysqld.cnf
bind-address = 0.0.0.0

# Restart MySQL
sudo systemctl restart mysql
```

### 1.2 Redis Configuration

Edit `/etc/redis/redis.conf`:
```conf
bind 0.0.0.0
requirepass YOUR_REDIS_PASSWORD
```

```bash
sudo systemctl restart redis
```

### 1.3 Firewall (Storage Server)

```bash
# Allow MySQL from app server
sudo ufw allow from APP_SERVER_IP to any port 3306

# Allow Redis from app server  
sudo ufw allow from APP_SERVER_IP to any port 6379
```

## Step 2: Prepare Application Server

### 2.1 Install Docker

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Add user to docker group
sudo usermod -aG docker $USER

# Install Docker Compose
sudo apt install docker-compose-plugin -y

# Verify installation
docker --version
docker compose version
```

### 2.2 Clone and Configure Nexus

```bash
# Clone repository
git clone YOUR_REPO_URL /opt/nexus
cd /opt/nexus

# Create environment files
cp nexus-core/.env.production.example nexus-core/.env
cp deploy/nexus_ui.env.example nexus_ui/.env.production

# Edit both files with your actual values (see next section)
```

### 2.3 Configure Environment Variables

Edit `nexus-core/.env`:
```bash
# Database (Storage Server)
NEXUS_DB_HOST=STORAGE_SERVER_IP
NEXUS_DB_PORT=3306
NEXUS_DB_USER=nexus
NEXUS_DB_PASSWORD=YOUR_MYSQL_PASSWORD
NEXUS_DB_NAME=nexus

# Redis (Storage Server)
NEXUS_REDIS_HOST=STORAGE_SERVER_IP
NEXUS_REDIS_PORT=6379
NEXUS_REDIS_PASSWORD=YOUR_REDIS_PASSWORD
NEXUS_REDIS_DB=0

# Security (generate new keys!)
NEXUS_JWT_SECRET=$(openssl rand -base64 32)
NEXUS_ENCRYPTION_KEY=$(openssl rand -base64 32)

# API
GIN_MODE=release
PORT=8080
```

Edit `nexus_ui/.env.production`:
```bash
NEXT_PUBLIC_API_URL=http://nexus-api:8080
DATABASE_URL="mysql://nexus:YOUR_PASSWORD@STORAGE_SERVER_IP:3306/nexus"
```

## Step 3: Run Database Migrations (Prisma)

Prisma manages database schema. You must run migrations **before** starting Docker containers.

### Option A: Run from Application Server (Recommended)

```bash
# SSH into application server
cd /opt/nexus/nexus_ui

# Install Node.js if not installed
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt-get install -y nodejs

# Install dependencies
npm install

# Set DATABASE_URL and run migrations
export DATABASE_URL="mysql://nexus:YOUR_PASSWORD@STORAGE_SERVER_IP:3306/nexus"
npx prisma migrate deploy ================= HEREEE

# Verify migration was successful
npx prisma migrate status

# Clean up (Docker will handle dependencies)
rm -rf node_modules
```

### Option B: Run from Local Machine

If your local machine can reach the storage server:

```bash
cd nexus_ui

# Point to production database
export DATABASE_URL="mysql://nexus:YOUR_PASSWORD@STORAGE_SERVER_IP:3306/nexus"

# Run migrations
npx prisma migrate deploy
```

### Migration Commands Reference

| Command | Purpose | When to Use |
|---------|---------|-------------|
| `npx prisma migrate deploy` | Apply pending migrations | **Production deployment** |
| `npx prisma migrate status` | Check migration status | Verify migrations applied |
| `npx prisma db push` | Push schema without migration files | Development/quick sync |
| `npx prisma migrate reset` | Reset database (DANGEROUS) | Never in production |

### Create Admin User

After migrations, create your first admin user:

```bash
# Connect to MySQL on storage server
mysql -h STORAGE_SERVER_IP -u nexus -p nexus

# Insert admin user (password: use bcrypt hash)
# Generate hash: https://bcrypt-generator.com/
INSERT INTO users (email, password_hash, name, role, is_active, created_at, updated_at) 
VALUES ('admin@yourcompany.com', '$2a$10$YOUR_BCRYPT_HASH', 'Admin', 'admin', 1, NOW(), NOW());
```

Or use the password hash tool:
```bash
cd /opt/nexus/nexus-core
go run tools/hash_password.go YOUR_PASSWORD
```

### Troubleshooting Migrations

**Error: Can't connect to MySQL server**
```bash
# Verify connection from app server
mysql -h STORAGE_IP -u nexus -p

# Check firewall allows 3306
sudo ufw status
```

**Error: Access denied for user**
```bash
# On storage server, verify user has correct permissions
mysql -u root -p
SHOW GRANTS FOR 'nexus'@'%';
```

**Error: Database doesn't exist**
```sql
-- On storage server
CREATE DATABASE nexus CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```


## Step 4: Build and Deploy

```bash
cd /opt/nexus

# Build and start all services
docker compose up -d --build

# Check status
docker compose ps

# View logs
docker compose logs -f
```

## Step 5: Verify Deployment

```bash
# Check API health
curl http://localhost:8080/health

# Check UI
curl http://localhost:3000

# Check via Nginx
curl http://localhost/api/health
curl http://localhost
```

## Management Commands

### View Logs
```bash
# All services
docker compose logs -f

# Specific service
docker compose logs -f nexus-api
docker compose logs -f nexus-worker
docker compose logs -f nexus-ui
```

### Restart Services
```bash
# All
docker compose restart

# Specific
docker compose restart nexus-worker
```

### Update Deployment
```bash
cd /opt/nexus
git pull
docker compose up -d --build
```

### Stop All Services
```bash
docker compose down
```

## Troubleshooting

### Container won't start
```bash
# Check logs
docker compose logs nexus-api

# Check if port is in use
sudo lsof -i :8080
```

### Database connection failed
```bash
# Test from app server
mysql -h STORAGE_IP -u nexus -p nexus

# Check firewall
sudo ufw status
```

### Redis connection failed
```bash
# Test from app server
redis-cli -h STORAGE_IP -p 6379 -a YOUR_PASSWORD ping
```

## Security Checklist

- [ ] Change default passwords
- [ ] Generate new JWT_SECRET and ENCRYPTION_KEY
- [ ] Restrict MySQL/Redis to private network only
- [ ] Enable UFW firewall on both servers
- [ ] Keep Docker images updated
- [ ] Set up log rotation
