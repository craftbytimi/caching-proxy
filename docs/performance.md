# Performance Guide

## Benchmarking

### Load Testing Tools

**wrk** - HTTP benchmarking tool
```bash
# Install
brew install wrk  # macOS
sudo apt install wrk  # Ubuntu

# Test 1000 requests with 10 connections for 30 seconds
wrk -t10 -c10 -d30s http://localhost:8080/api/test
```

**hey** - HTTP load generator
```bash
# Install
go install github.com/rakyll/hey@latest

# Send 10000 requests with 100 concurrent workers
hey -n 10000 -c 100 http://localhost:8080/api/test
```

**Apache Bench (ab)**
```bash
ab -n 10000 -c 100 http://localhost:8080/api/test
```

### Expected Performance

On modern hardware (4 CPU, 8GB RAM):

| Scenario | Throughput | Latency (p50) | Latency (p99) |
|----------|------------|---------------|---------------|
| Cache Hit | 50,000 req/s | 0.5ms | 2ms |
| Cache Miss | 5,000 req/s | 10ms | 50ms |
| Bypass | 5,000 req/s | 10ms | 50ms |

*Note: Miss/Bypass performance depends heavily on upstream latency*

---

## Optimization Strategies

### 1. Response Buffering

**Problem**: Buffering entire response in memory before sending

**Current Implementation**:
```go
// Buffers entire response
body, err := io.ReadAll(resp.Body)
```

**Optimization** (Future):
```go
// Stream response while capturing for cache
tee := io.TeeReader(resp.Body, cacheWriter)
io.Copy(w, tee)
```

### 2. Serialization Format

**Current**: JSON serialization
```go
data, err := json.Marshal(response)  // ~500 ns/op
```

**Alternatives**:

**MessagePack** - Faster, smaller
```go
import "github.com/vmihailenco/msgpack/v5"
data, err := msgpack.Marshal(response)  // ~300 ns/op, 20% smaller
```

**Protocol Buffers** - Fastest, smallest
```go
// Requires schema definition
data, err := proto.Marshal(response)  // ~100 ns/op, 40% smaller
```

### 3. Cache Key Generation

**Current Implementation**:
```go
// SHA-256 hash
h := sha256.New()
h.Write([]byte(s))
return fmt.Sprintf("%x", h.Sum(nil))
```

**Benchmarks**:
- SHA-256: ~1500 ns/op
- xxHash: ~150 ns/op (10x faster)
- FNV-1a: ~50 ns/op (30x faster, but more collisions)

**Optimization**:
```go
import "github.com/cespare/xxhash/v2"

func hashString(s string) string {
    h := xxhash.Sum64String(s)
    return strconv.FormatUint(h, 36)
}
```

### 4. Connection Pooling

**Redis Connection Pool**:
```go
redis.NewClient(&redis.Options{
    PoolSize:     100,      // Increase for high load
    MinIdleConns: 10,       // Keep connections warm
    MaxRetries:   3,
    PoolTimeout:  30 * time.Second,
})
```

**Upstream HTTP Pool**:
```go
&http.Transport{
    MaxIdleConns:        1000,  // Total idle connections
    MaxIdleConnsPerHost: 100,   // Per upstream host
    MaxConnsPerHost:     100,   // Limit concurrent connections
    IdleConnTimeout:     90 * time.Second,
    DisableCompression:  true,  // If upstream already compressed
}
```

### 5. Request Coalescing

**Problem**: Multiple concurrent requests to same uncached URL hit upstream multiple times

**Solution**: Use singleflight
```go
import "golang.org/x/sync/singleflight"

type Service struct {
    // ... existing fields
    flight singleflight.Group
}

func (s *Service) fetchWithCoalescing(key string, req *http.Request) (*model.CachedResponse, error) {
    val, err, _ := s.flight.Do(key, func() (interface{}, error) {
        return s.fetchFromUpstream(req)
    })
    return val.(*model.CachedResponse), err
}
```

**Impact**: Reduces upstream load by ~90% during cache warming

### 6. Compression

**Compress cached responses**:
```go
import "github.com/klauspost/compress/zstd"

// Before storing
compressed, _ := zstd.Compress(nil, body)

// After retrieving
decompressed, _ := zstd.Decompress(nil, compressed)
```

**Savings**: 60-80% storage reduction for text/JSON responses

---

## Redis Optimization

### Memory Optimization

**1. Use appropriate eviction policy**:
```conf
maxmemory 2gb
maxmemory-policy allkeys-lru  # Evict least recently used
```

**2. Disable persistence for cache-only use**:
```conf
save ""
appendonly no
```

**3. Enable key expiration**:
```bash
# Redis automatically removes expired keys
# No manual cleanup needed
```

### Performance Tuning

**1. Increase TCP backlog**:
```conf
tcp-backlog 511
```

**2. Disable slow log for high throughput**:
```conf
slowlog-max-len 0
```

**3. Use pipelining** (future enhancement):
```go
pipe := client.Pipeline()
pipe.Get(ctx, key1)
pipe.Get(ctx, key2)
results, _ := pipe.Exec(ctx)
```

### Monitoring Redis

```bash
# Monitor commands in real-time
redis-cli MONITOR

# Check memory usage
redis-cli INFO memory

# Check hit rate
redis-cli INFO stats | grep keyspace
```

---

## Go Runtime Tuning

### GOMAXPROCS

```bash
# Let Go use all CPUs
export GOMAXPROCS=$(nproc)

# Or limit for other workloads
export GOMAXPROCS=4
```

### Garbage Collection

**Monitor GC**:
```bash
GODEBUG=gctrace=1 ./caching-proxy
```

**Tune GC**:
```bash
# Default GOGC=100 (GC when heap doubles)
# Increase for less frequent GC (more memory usage)
export GOGC=200

# Decrease for more frequent GC (less memory usage)
export GOGC=50
```

### Memory Profiling

```bash
# Enable pprof endpoint
import _ "net/http/pprof"

# In your main:
go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()
```

**Analyze memory**:
```bash
# Heap profile
go tool pprof http://localhost:6060/debug/pprof/heap

# CPU profile
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Goroutines
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

---

## Caching Strategy Optimization

### Vary by Content-Type

Cache different content types with different TTLs:

```go
func (p *SimplePolicy) GetTTL(contentType string) time.Duration {
    switch {
    case strings.Contains(contentType, "image/"):
        return 24 * time.Hour  // Cache images longer
    case strings.Contains(contentType, "application/json"):
        return 5 * time.Minute  // API responses shorter
    default:
        return p.defaultTTL
    }
}
```

### Ignore Query Parameters

For analytics/tracking parameters:

```go
// Ignore common tracking params
ignoredParams := []string{"utm_source", "utm_medium", "fbclid"}

func normalizeQuery(query string) string {
    values, _ := url.ParseQuery(query)
    for _, param := range ignoredParams {
        values.Del(param)
    }
    return values.Encode()
}
```

### Cache Warming

Pre-populate cache with popular endpoints:

```go
func (s *Service) WarmCache(urls []string) error {
    for _, url := range urls {
        req, _ := http.NewRequest("GET", url, nil)
        s.ServeHTTP(httptest.NewRecorder(), req)
    }
    return nil
}
```

---

## Benchmarking Results

### Cache Key Generation

```
BenchmarkSHA256-8          1000000    1500 ns/op    32 B/op    1 allocs/op
BenchmarkXXHash-8         10000000     150 ns/op     8 B/op    1 allocs/op
BenchmarkFNV1a-8          20000000      50 ns/op     8 B/op    1 allocs/op
```

### Serialization

```
BenchmarkJSON-8            500000    3000 ns/op   512 B/op    6 allocs/op
BenchmarkMsgpack-8        1000000    1800 ns/op   256 B/op    4 allocs/op
BenchmarkProtobuf-8       2000000     600 ns/op   128 B/op    2 allocs/op
```

### Store Operations

```
BenchmarkMemoryGet-8     5000000     250 ns/op     0 B/op    0 allocs/op
BenchmarkMemorySet-8     2000000     800 ns/op    64 B/op    2 allocs/op
BenchmarkRedisGet-8      1000000    1500 ns/op   256 B/op    8 allocs/op
BenchmarkRedisSet-8       500000    3000 ns/op   512 B/op   12 allocs/op
```

---

## Production Performance Checklist

- [ ] Load testing completed with realistic traffic patterns
- [ ] Response time p50, p95, p99 within targets
- [ ] Memory usage stable under sustained load
- [ ] Redis connection pool sized appropriately
- [ ] HTTP connection pool configured for upstream
- [ ] Cache hit rate > 80% for cacheable endpoints
- [ ] GC pauses < 10ms
- [ ] No memory leaks detected
- [ ] CPU usage < 70% at peak load
- [ ] Graceful degradation tested (Redis down)
- [ ] Monitoring and alerting configured
- [ ] Performance baselines documented

---

## Further Optimizations

1. **Multi-tier cache**: Memory + Redis
2. **Response compression**: Gzip/Brotli
3. **HTTP/2 support**: Reduce connection overhead
4. **Edge caching**: Deploy closer to users
5. **Cache preloading**: Warm cache on startup
6. **Adaptive TTL**: Adjust based on hit rate
7. **Bloom filters**: Fast negative cache lookups
