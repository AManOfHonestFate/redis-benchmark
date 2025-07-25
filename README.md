# Redis Cluster with Docker (Dynamic & Production Ready)

This setup creates a Redis cluster using Docker Compose, with dynamic node count, security, and production best practices.

## Prerequisites
- Docker
- Docker Compose

## Setup

### 1. Configure Cluster

Edit the `.env` file to set the number of nodes and password:

```
REDIS_NODES=6
REDIS_PASSWORD=yourStrongPasswordHere
```

### 2. Generate Compose and Scripts

Run the generator script:

```
./generate-cluster.sh
```

This creates `docker-compose.generated.yml` and `init-cluster.sh` for your cluster size.

### 3. Build and Start the Cluster

```
docker-compose -f docker-compose.generated.yml up -d --build
```

### 4. Initialize the Cluster

```
./init-cluster.sh
```

### 5. Connect to the Cluster

```
redis-cli -a $REDIS_PASSWORD -c -p 7001
```

### 6. Monitoring

Prometheus-compatible metrics are available at `localhost:9121` via the `redis-exporter` service.

### 7. Backup & Restore

To backup:
```
./backup.sh
```
To restore:
```
./restore.sh <backup_file>
```

### 8. Security
- Password auth is enabled by default.
- TLS config is included but commented out; see `redis.conf` and `Dockerfile`.
- Ports are mapped to localhost by default; restrict as needed for production.

### 9. Scaling
- Change `REDIS_NODES` in `.env` and re-run `generate-cluster.sh`.
- Bring down the cluster and remove volumes if reducing nodes:
```
docker-compose -f docker-compose.generated.yml down -v
```

---

For advanced production hardening (TLS, Docker secrets, etc.), see comments in config files.
