**Build Plan**

This is the order I’d actually build it in for `Go + stdlib + modular monolith`, keeping the MVP small but clean.

**V1 Scope**
- Start the proxy from CLI: `caching-proxy --port 3000 --upstream http://example.com --ttl 60 --redis redis://localhost:6379`
- Forward requests to one configured upstream.
- Cache only `GET` requests.
- Cache only `200 OK` responses.
- Use Redis TTL cache as the backend.
- Return `X-Cache: HIT` or `X-Cache: MISS`.
- Add `/healthz`.
- Skip file-based persistence, `ETag`, `Cache-Control` parsing, and advanced invalidation for v1.

**Target Package Layout**
```text
cmd/caching-proxy/main.go
internal/config/config.go
internal/model/model.go
internal/server/server.go
internal/proxy/service.go
internal/cache/key.go
internal/cache/policy.go
internal/store/memory/store.go
internal/upstream/client.go
internal/observability/logger.go
```

**Phase 1: Bootstrap The Project**
- Goal: get a runnable Go app with clean package boundaries.
- Tasks: initialize `go.mod`; create package skeleton; add `main.go`; wire a basic logger; add a no-op `/healthz` server.
- Done when: `go run ./cmd/caching-proxy --port 3000 --upstream http://example.com` starts and `/healthz` returns `200`.

**Phase 2: Define Core Contracts**
- Goal: lock the internal shape before implementation spreads.
- Tasks: define `Config`, `CachedResponse`, `CacheStore`, `UpstreamClient`, and maybe a small `Clock` abstraction for TTL tests.
- Done when: all packages depend on interfaces/types instead of concrete cross-package knowledge.

**Phase 3: CLI And Config**
- Goal: make startup predictable and validate inputs early.
- Tasks: parse flags with `flag`; validate `port`, `upstream`, and `ttl`; normalize upstream URL; set sensible defaults; print startup config in logs.
- Done when: invalid config fails fast with clear errors and valid config boots cleanly.

**Phase 4: Build A Pass-Through Proxy First**
- Goal: get the full request path working before adding cache complexity.
- Tasks: implement `internal/upstream/client.go` using `http.Client`; copy request method, path, query, body, and safe headers; strip hop-by-hop headers; return upstream status, headers, and body.
- Done when: the proxy correctly forwards requests and mirrors upstream responses without caching.

**Phase 5: Implement The Redis Store**
- Goal: add a safe, testable cache backend using Redis.
- Tasks: integrate Redis using `github.com/redis/go-redis/v9`; store `CachedResponse` with TTL; serialize/deserialize responses for storage; ensure atomicity for cache operations.
- Done when: `Get`, `Set`, `Delete`, and `Clear` behave correctly with Redis and TTL is enforced.

**Phase 6: Implement Cache Key And Policy**
- Goal: centralize all cache decisions in one place.
- Tasks: create a cache key from method + normalized path + query; decide whether to vary by headers; add `ShouldReadFromCache` and `ShouldStoreResponse`; skip caching if request has `Authorization` or sensitive cookies unless you explicitly want that behavior.
- Done when: cacheability rules are isolated in `internal/cache` and easy to unit test.

**Phase 7: Integrate The Proxy Service**
- Goal: orchestrate request -> cache -> upstream -> response.
- Tasks: in `internal/proxy/service.go`, check cache first; on hit return cached response; on miss call upstream; evaluate policy; store cacheable responses; attach `X-Cache`.
- Done when: repeated `GET` requests return `MISS` then `HIT`, and non-cacheable requests always bypass cache.

**Phase 8: Harden The HTTP Server**
- Goal: make the server safe enough for real local use.
- Tasks: add server timeouts (`ReadHeaderTimeout`, `IdleTimeout`, `WriteTimeout` as appropriate); add upstream client timeout; cap max cached body size to avoid memory blowups; add graceful shutdown handling.
- Done when: the server handles slow or large requests predictably and exits cleanly.

**Phase 9: Add Observability**
- Goal: make behavior easy to debug.
- Tasks: log request method, path, status, duration, cache status, and upstream target with `slog`; log startup config; log upstream errors clearly.
- Done when: one request log line is enough to tell whether a request hit cache, missed, failed upstream, or was not cacheable.

**Phase 10: Write Tests In Layers**
- Goal: verify correctness without needing a real external server.
- Tasks: unit test cache key generation, policy decisions, and memory store TTL behavior; integration test proxy flow with `httptest`; test `MISS -> HIT`, TTL expiry, query-string key separation, header copying, non-GET bypass, non-200 bypass, and upstream failure behavior.
- Done when: `go test ./...` covers both happy paths and the main edge cases.

**Phase 11: Documentation And Developer Experience**
- Goal: make the project easy to run and explain.
- Tasks: write a short `README` with usage, flags, example requests, and `X-Cache` behavior; document the modular-monolith structure and v1 non-goals.
- Done when: someone can clone, run, and understand the design in a few minutes.

**Recommended Build Order**
1. Boot the app and `/healthz`.
2. Implement pass-through proxy without cache.
3. Add Redis store.
4. Add cache key/policy.
5. Wire cache into proxy service.
6. Add tests.
7. Harden with timeouts, limits, and logs.
8. Document it.

**Definition Of Done For V1**
- Proxy starts from CLI.
- Requests forward correctly to upstream.
- Repeated cacheable `GET` requests return `MISS` then `HIT`.
- Expired entries are not served (enforced by Redis TTL).
- Logs clearly show request outcome.
- Tests pass with `go test ./...`.
- README explains how to run and verify behavior.

**Stretch Backlog After V1**
- LRU eviction in addition to TTL.
- background cleanup goroutine for expired entries.
- respect `Cache-Control` and `ETag`.
- file-based store.
- singleflight request coalescing to avoid duplicate upstream fetches on concurrent misses.
- admin endpoint or CLI command to clear cache.
- metrics endpoint for hit rate and cache size.

**Redis Setup**

You must have a running Redis instance. By default, the app expects Redis at `redis://localhost:6379`. You can override this with the `--redis` flag.

**Example usage:**

```
caching-proxy --port 3000 --upstream http://example.com --ttl 60 --redis redis://localhost:6379
```

The cache backend uses Redis for storing cached responses with TTL. Ensure Redis is running before starting the proxy.

If you want, I can turn this next into a concrete task board with exact files, functions, and implementation order for each file.
