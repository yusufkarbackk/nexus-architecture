# Nexus Podman Deployment Guide

Panduan lengkap untuk menjalankan Nexus menggunakan Podman (alternatif rootless untuk Docker).

---

## Table of Contents
- [Prerequisites](#prerequisites)
- [Windows Setup](#windows-setup)
- [Local Development](#local-development)
- [Production Deployment](#production-deployment)
- [Command Reference](#command-reference)
- [Troubleshooting](#troubleshooting)

---

## Prerequisites

### Windows
- **Podman Desktop** - Download dari [podman-desktop.io](https://podman-desktop.io)
- **podman-compose** - Install via pip:
  ```powershell
  pip install podman-compose
  ```

### Linux (Ubuntu/Debian)
```bash
sudo apt-get update
sudo apt-get install -y podman podman-compose
```

### Linux (RHEL/CentOS/Fedora)
```bash
sudo dnf install podman podman-compose
```

### Verifikasi Instalasi
```bash
podman --version
# Output: podman version 5.x.x

podman-compose --version
# Output: podman-compose version x.x.x
```

---

## Windows Setup

### 1. Inisialisasi Podman Machine

Podman Desktop biasanya sudah meng-handle ini, tapi jika perlu manual:

```powershell
# Buat dan start machine
podman machine init
podman machine start

# Verifikasi
podman info
```

### 2. (Opsional) Alias Docker ke Podman

Untuk kompatibilitas dengan skrip yang menggunakan perintah `docker`:

```powershell
# Di PowerShell profile (notepad $PROFILE)
function docker { podman $args }
function docker-compose { podman-compose $args }
```

---

## Local Development

### Step 1: Clone Repository

```bash
git clone --recurse-submodules https://github.com/YOUR_ORG/nexus-architecture.git
cd nexus-architecture
```

### Step 2: Create Network & Infrastructure

```bash
# Create network
podman network create nexus-infra_default

# Start MySQL
podman run -d \
  --name nexus_mysql_local \
  --network nexus-infra_default \
  -e MYSQL_ROOT_PASSWORD=root \
  -e MYSQL_DATABASE=nexus_db \
  -e MYSQL_USER=demo_user \
  -e MYSQL_PASSWORD=demo_password \
  -p 3309:3306 \
  mysql:8.0

# Start Redis
podman run -d \
  --name nexus_redis_local \
  --network nexus-infra_default \
  -p 6381:6379 \
  redis:alpine
```

### Step 3: Create Environment Files

Sama seperti Docker - lihat [DOCKER_DEPLOYMENT.md](./DOCKER_DEPLOYMENT.md#step-3-create-environment-files)

### Step 4: Start Nexus Services

```bash
# Build dan jalankan semua services
podman-compose -f docker-compose.local.yml up --build

# Atau jalankan di background
podman-compose -f docker-compose.local.yml up -d --build
```

### Step 5: Access Services

| Service | URL |
|---------|-----|
| Nexus UI | http://localhost:3000 |
| Nexus API | http://localhost:8080 |
| Nginx | http://localhost |

### Step 6: Stop Services

```bash
# Stop compose services
podman-compose -f docker-compose.local.yml down

# Stop infrastructure (jika diperlukan)
podman stop nexus_mysql_local nexus_redis_local
podman rm nexus_mysql_local nexus_redis_local
```

---

## Production Deployment

### Step 1: Server Setup (Linux)

```bash
# Install Podman
sudo apt-get update
sudo apt-get install -y podman podman-compose

# Enable lingering untuk user (agar services tetap jalan setelah logout)
sudo loginctl enable-linger $USER
```

### Step 2: Clone & Configure

```bash
cd /opt
git clone --recurse-submodules https://github.com/YOUR_ORG/nexus-architecture.git
cd nexus-architecture

# Setup environment files (sama seperti Docker)
# Edit nexus-core/.env dan nexus-core/config.yml
```

### Step 3: Build and Deploy

```bash
# Build semua images
podman-compose build --no-cache

# Start services
podman-compose up -d
```

### Step 4: Verify

```bash
# Check status
podman-compose ps

# Check logs
podman-compose logs -f

# Health check
curl http://localhost:8080/health
```

### Step 5: Setup Systemd (Auto-Start on Boot)

```bash
# Generate systemd unit files untuk setiap container
mkdir -p ~/.config/systemd/user

# Generate untuk nexus-api
podman generate systemd --new --name nexus-api > ~/.config/systemd/user/nexus-api.service

# Enable dan start
systemctl --user daemon-reload
systemctl --user enable nexus-api.service
systemctl --user start nexus-api.service
```

Atau gunakan quadlet (Podman 4.4+) untuk integrasi systemd yang lebih baik.

---

## Command Reference

### Podman vs Docker Commands

| Docker | Podman | Keterangan |
|--------|--------|------------|
| `docker build` | `podman build` | Build image |
| `docker run` | `podman run` | Run container |
| `docker-compose up` | `podman-compose up` | Start compose |
| `docker-compose down` | `podman-compose down` | Stop compose |
| `docker ps` | `podman ps` | List containers |
| `docker exec -it` | `podman exec -it` | Execute in container |
| `docker logs` | `podman logs` | View logs |
| `docker network ls` | `podman network ls` | List networks |
| `docker system prune` | `podman system prune` | Cleanup |

### Useful Commands

```bash
# List all containers (including stopped)
podman ps -a

# View container logs
podman logs -f nexus-api

# Enter container shell
podman exec -it nexus-api sh

# Inspect container
podman inspect nexus-api

# Check resource usage
podman stats

# Clean up unused images/containers
podman system prune -a
```

---

## Troubleshooting

### Issue: Permission Denied (Rootless)

```bash
# Migrate storage jika perlu
podman system migrate

# Atau reset storage
podman system reset
```

### Issue: Network DNS Not Resolving

```bash
# Gunakan slirp4netns untuk networking
podman run --network slirp4netns:port_handler=slirp4netns ...
```

### Issue: Volume Mount Permission

Tambahkan `:Z` pada volume mounts untuk SELinux relabeling:

```yaml
volumes:
  - ./config.yml:/app/config.yml:ro,Z
```

### Issue: Podman Machine Not Running (Windows)

```powershell
# Check status
podman machine ls

# Start machine
podman machine start

# Restart jika bermasalah
podman machine stop
podman machine start
```

### Issue: Container Tidak Bisa Connect ke Service Lain

Pastikan semua container dalam network yang sama:

```bash
# Check network
podman network inspect nexus-infra_default

# Pastikan container dalam network
podman inspect nexus-api | grep NetworkMode
```

### Issue: Build Gagal Karena SSL/TLS

Untuk corporate proxy atau SSL issues:

```bash
# Build dengan skip SSL verification
podman build --tls-verify=false -t myimage .
```

---

## Notes

> **💡 Tip**: Semua Dockerfile yang ada sudah kompatibel 100% dengan Podman. Tidak perlu modifikasi apapun.

> **⚠️ Penting**: Jika menggunakan Windows, pastikan Podman Machine sudah running sebelum menjalankan perintah podman.

---

## Migration from Docker

Jika sebelumnya menggunakan Docker:

1. **Stop Docker services**: `docker-compose down`
2. **Install Podman** (lihat Prerequisites)
3. **Start dengan Podman**: `podman-compose -f docker-compose.local.yml up --build`

Tidak perlu mengubah Dockerfile atau docker-compose files - semuanya kompatibel!
