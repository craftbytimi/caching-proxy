# Deployment Guide

## Prerequisites

- Go 1.21 or later
- Redis 6.0 or later
- Docker (optional, for Redis)

## Local Development

### 1. Install Dependencies

```bash
make install
```

### 2. Start Redis

Using Docker:
```bash
make redis
```

Or manually:
```bash
redis-server
```

### 3. Configure Environment

Copy the example environment file:
```bash
cp .env.example .env
```

Edit `.env` with your configuration:
```env
PORT=8080
UPSTREAM_URL=http://example.com
CACHE_TTL=60
REDIS_ADDR=localhost:6379
```

### 4. Run the Proxy

Using Make:
```bash
make run
```

Or directly:
```bash
go run ./cmd/proxy --port 8080 --upstream http://example.com --ttl 60
```

With live reload (requires Air):
```bash
make dev
```

---

## Docker Deployment

### Build Docker Image

```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /caching-proxy ./cmd/proxy

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /caching-proxy .

EXPOSE 8080
CMD ["./caching-proxy"]
```

Build and run:
```bash
docker build -t caching-proxy .
docker run -p 8080:8080 \
  -e UPSTREAM_URL=http://example.com \
  -e REDIS_ADDR=redis:6379 \
  caching-proxy
```

### Docker Compose

```yaml
# docker-compose.yml
version: '3.8'

services:
  proxy:
    build: .
    ports:
      - "8080:8080"
    environment:
      - UPSTREAM_URL=http://example.com
      - REDIS_ADDR=redis:6379
      - CACHE_TTL=60
      - LOG_LEVEL=info
    depends_on:
      - redis
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data
    restart: unless-stopped

volumes:
  redis-data:
```

Start services:
```bash
docker-compose up -d
```

---

## Production Deployment

### Systemd Service

Create `/etc/systemd/system/caching-proxy.service`:

```ini
[Unit]
Description=Caching Proxy Server
After=network.target redis.service

[Service]
Type=simple
User=proxy
Group=proxy
WorkingDirectory=/opt/caching-proxy
ExecStart=/opt/caching-proxy/bin/caching-proxy \
  --port 8080 \
  --upstream http://api.example.com \
  --ttl 300 \
  --redis localhost:6379
Restart=always
RestartSec=5

# Security
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/log/caching-proxy

# Limits
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl enable caching-proxy
sudo systemctl start caching-proxy
sudo systemctl status caching-proxy
```

View logs:
```bash
sudo journalctl -u caching-proxy -f
```

---

## Redis Configuration

### Production Redis Setup

Create `/etc/redis/redis.conf`:

```conf
# Basic
bind 127.0.0.1
port 6379
daemonize yes
pidfile /var/run/redis/redis.pid

# Persistence
save 900 1
save 300 10
save 60 10000
dir /var/lib/redis

# Memory
maxmemory 2gb
maxmemory-policy allkeys-lru

# Logging
loglevel notice
logfile /var/log/redis/redis.log

# Performance
tcp-backlog 511
timeout 0
tcp-keepalive 300
```

### Redis Cluster (Optional)

For high availability, use Redis Sentinel or Redis Cluster:

```bash
# Redis Sentinel example
redis-sentinel /etc/redis/sentinel.conf
```

Update proxy configuration:
```bash
REDIS_ADDR=sentinel://localhost:26379
```

---

## Monitoring

### Health Checks

Configure your load balancer or monitoring tool to check:

```bash
curl http://localhost:8080/healthz
```

Expected response: `200 OK` with body `ok`

### Metrics (Future Enhancement)

Expose Prometheus metrics at `/metrics`:

- `cache_hits_total` - Total cache hits
- `cache_misses_total` - Total cache misses
- `request_duration_seconds` - Request latency histogram
- `upstream_duration_seconds` - Upstream latency histogram
- `cache_size_bytes` - Current cache size

Prometheus scrape config:
```yaml
scrape_configs:
  - job_name: 'caching-proxy'
    static_configs:
      - targets: ['localhost:8080']
```

### Logging

Logs are output as JSON to stdout. Use a log aggregator like:

- **Logstash** + Elasticsearch + Kibana
- **Grafana Loki**
- **CloudWatch Logs** (AWS)
- **Stackdriver** (GCP)

Example log entry:
```json
{
  "time": "2026-03-26T10:30:45Z",
  "level": "INFO",
  "msg": "request completed",
  "method": "GET",
  "path": "/api/users",
  "status": 200,
  "duration_ms": 45,
  "request_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

---

## Performance Tuning

### Redis Tuning

1. **Memory Policy**: Use `allkeys-lru` for automatic eviction
2. **Persistence**: Disable if cache is ephemeral (`save ""`)
3. **Connection Pool**: Increase `MaxIdleConns` in client
4. **TCP Buffer**: Increase `tcp-backlog` in redis.conf

### Proxy Tuning

1. **Timeouts**: Adjust based on upstream latency
   ```bash
   --timeout 30  # 30 second upstream timeout
   ```

2. **Max Body Size**: Limit large responses
   ```bash
   --max-body 50  # 50MB max cacheable size
   ```

3. **TTL**: Tune based on content freshness requirements
   ```bash
   --ttl 3600  # 1 hour cache
   ```

4. **Go Runtime**:
   ```bash
   export GOMAXPROCS=8
   export GOGC=100
   ```

### Load Balancing

Use multiple proxy instances behind a load balancer:

```
           ┌─── Proxy Instance 1 ───┐
           │                         │
Client ──→ LB ──→ Proxy Instance 2 ──→ Redis ──→ Upstream
           │                         │
           └─── Proxy Instance 3 ───┘
```

All instances share the same Redis cache.

---

## Security

### Network Security

1. **Use HTTPS** for client connections (add TLS termination)
2. **Secure Redis** with password and bind to localhost
3. **Firewall Rules** to restrict access

### Redis Authentication

```conf
requirepass your-strong-password
```

Update proxy config:
```bash
REDIS_PASSWORD=your-strong-password
```

### Rate Limiting (Future Enhancement)

Add rate limiting middleware to prevent abuse.

---

## Backup and Recovery

### Redis Backup

**RDB Snapshots**:
```bash
# Manual backup
redis-cli BGSAVE

# Copy snapshot
cp /var/lib/redis/dump.rdb /backup/dump-$(date +%Y%m%d).rdb
```

**AOF (Append-Only File)**:
```conf
appendonly yes
appendfsync everysec
```

### Disaster Recovery

1. Cache data is ephemeral - OK to lose
2. Proxy will fallback to upstream if Redis is down
3. No persistent state in proxy itself

---

## Troubleshooting

### Proxy Won't Start

1. Check if port is available:
   ```bash
   lsof -i :8080
   ```

2. Verify Redis connection:
   ```bash
   redis-cli ping
   ```

3. Check logs:
   ```bash
   journalctl -u caching-proxy -n 50
   ```

### High Memory Usage

1. Check Redis memory:
   ```bash
   redis-cli INFO memory
   ```

2. Verify max body size configuration
3. Check for memory leaks with pprof

### Poor Cache Hit Rate

1. Check cache policy configuration
2. Verify TTL is not too short
3. Monitor cache keys:
   ```bash
   redis-cli KEYS '*' | wc -l
   ```

4. Check for high request variance (query params)

---

## Scaling Checklist

- [ ] Redis cluster or Sentinel for HA
- [ ] Multiple proxy instances behind load balancer
- [ ] Monitoring and alerting configured
- [ ] Log aggregation set up
- [ ] Health checks configured
- [ ] Backup strategy defined
- [ ] Performance testing completed
- [ ] Documentation updated
