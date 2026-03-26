# Quick Start Guide

Get the caching proxy up and running in 5 minutes.

> 📚 **Documentation Navigation**
> - 🏠 [README.md](README.md) - Main documentation hub
> - 📊 [PROJECT_STATUS.md](PROJECT_STATUS.md) - Implementation progress
> - 🏗️ [ARCHITECTURE.md](ARCHITECTURE.md) - Detailed architecture & development plan
> - 📦 [SCAFFOLD_SUMMARY.md](SCAFFOLD_SUMMARY.md) - Project scaffold overview

## Prerequisites

- Go 1.21+ installed
- Redis installed or Docker available

## Steps

### 1. Clone and Navigate

```bash
cd /home/eyiowuawi/Documents/projects/caching-proxy
```

### 2. Install Dependencies

```bash
make install
```

### 3. Start Redis

**Option A: Using Docker (Recommended)**
```bash
make redis
```

**Option B: Local Redis**
```bash
redis-server
```

### 4. Configure Environment

```bash
cp .env.example .env
```

Edit `.env` with your settings:
```env
PORT=8080
UPSTREAM_URL=http://httpbin.org
CACHE_TTL=60
REDIS_ADDR=localhost:6379
LOG_LEVEL=info
```

### 5. Run the Proxy

**Note**: The main.go file needs to be updated first (see Phase 7 in PROJECT_STATUS.md)

Once updated, run:
```bash
make run
```

Or with CLI flags:
```bash
go run ./cmd/proxy \
  --port 8080 \
  --upstream http://httpbin.org \
  --ttl 60 \
  --redis localhost:6379 \
  --log-level info
```

### 6. Test It

**Health Check**:
```bash
curl http://localhost:8080/healthz
# Expected: ok
```

**Proxy Request (once implemented)**:
```bash
# First request - cache miss
curl -i http://localhost:8080/get
# X-Cache: MISS

# Second request - cache hit
curl -i http://localhost:8080/get
# X-Cache: HIT
```

## Development Workflow

### Run with Live Reload

```bash
make dev
```

Changes to `.go` files will automatically rebuild and restart the server.

### Run Tests

```bash
# All tests
make test

# Unit tests only
make test-unit

# Integration tests only
make test-integration
```

### Build Binary

```bash
make build
# Binary created at: bin/caching-proxy
```

## Common Commands

```bash
make help           # Show all available commands
make redis          # Start Redis container
make redis-stop     # Stop Redis container
make fmt            # Format code
make lint           # Run linter (requires golangci-lint)
make clean          # Clean build artifacts
```

## Troubleshooting

### Redis Connection Failed

**Error**: `could not connect to redis: connection refused`

**Solution**:
```bash
# Check if Redis is running
redis-cli ping

# If not, start it
make redis
```

### Port Already in Use

**Error**: `bind: address already in use`

**Solution**:
```bash
# Find process using port 8080
lsof -i :8080

# Kill it or use different port
go run ./cmd/proxy --port 8081 ...
```

### Module Errors

**Error**: `package github.com/... not found`

**Solution**:
```bash
go mod download
go mod tidy
```

## Next Steps

1. **Read [ARCHITECTURE.md](ARCHITECTURE.md)** to understand the project structure
2. **Check [PROJECT_STATUS.md](PROJECT_STATUS.md)** for implementation progress
3. **Review [docs/](docs/)** for detailed documentation:
   - [docs/api.md](docs/api.md) - API documentation
   - [docs/deployment.md](docs/deployment.md) - Deployment guide
   - [docs/performance.md](docs/performance.md) - Performance tuning
4. **Review [SCAFFOLD_SUMMARY.md](SCAFFOLD_SUMMARY.md)** for project overview
5. **Start implementing** following the phase plan in [ARCHITECTURE.md](ARCHITECTURE.md)

## Useful Resources

- [Go Documentation](https://go.dev/doc/)
- [Redis Go Client](https://redis.uptrace.dev/)
- [HTTP Caching RFC](https://www.rfc-editor.org/rfc/rfc9111.html)

## Getting Help

- Check the `docs/` directory for detailed guides
- Review test files for usage examples
- Open an issue for bugs or questions

---

**Happy Coding!** 🚀
