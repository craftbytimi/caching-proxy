# API Documentation

## Endpoints

### Proxy Endpoints

All requests to paths other than `/healthz` are proxied to the configured upstream server.

#### Example: Proxy Request

```bash
GET /api/users
Host: localhost:8080
```

**Response:**
```
HTTP/1.1 200 OK
X-Cache: MISS
Content-Type: application/json

[upstream response body]
```

**Cache Headers:**
- `X-Cache: HIT` - Response served from cache
- `X-Cache: MISS` - Response fetched from upstream and cached
- `X-Cache: BYPASS` - Request bypassed cache (non-cacheable)
- `X-Request-ID` - Unique request identifier for tracing

---

### Health Check

Check if the proxy and its dependencies are healthy.

#### Request

```
GET /healthz
```

#### Response (Healthy)

```
HTTP/1.1 200 OK
Content-Type: text/plain

ok
```

#### Response (Unhealthy)

```
HTTP/1.1 503 Service Unavailable
Content-Type: text/plain

store unhealthy: [error details]
```

---

## Cache Behavior

### Cacheable Requests

Requests are cached when ALL of the following conditions are met:

1. HTTP method is `GET`
2. No `Authorization` header present
3. No sensitive cookies (session, auth, token)

### Cacheable Responses

Responses are stored when ALL of the following conditions are met:

1. HTTP status code is `200 OK`
2. Response body size is within configured limit
3. No `Set-Cookie` header present
4. No `Cache-Control: no-store` header present

### Cache Keys

Cache keys are generated from:
- HTTP method (e.g., `GET`)
- Normalized URL (lowercase host, sorted query parameters)

Example cache key components:
```
GET + http://example.com/api/users?sort=name&page=1
→ SHA-256 hash
```

### TTL (Time To Live)

All cached entries expire after the configured TTL period. Default is 60 seconds.

---

## Error Responses

### 502 Bad Gateway

Returned when the upstream server is unreachable or returns an error.

```
HTTP/1.1 502 Bad Gateway
Content-Type: text/plain

Bad Gateway
```

### 503 Service Unavailable

Returned when the proxy itself is unhealthy (e.g., Redis connection lost).

```
HTTP/1.1 503 Service Unavailable
Content-Type: text/plain

store unhealthy: connection refused
```

---

## Request/Response Flow

```
1. Client sends request to proxy
2. Proxy checks if request is cacheable
   └─ If NO → Forward to upstream (BYPASS)
   └─ If YES → Continue to step 3
3. Proxy checks cache for matching key
   └─ If FOUND → Return cached response (HIT)
   └─ If NOT FOUND → Continue to step 4
4. Proxy forwards request to upstream
5. Proxy receives upstream response
6. Proxy checks if response is cacheable
   └─ If YES → Store in cache with TTL
   └─ If NO → Skip caching
7. Proxy returns response to client (MISS)
```

---

## Headers

### Request Headers Forwarded

All headers except hop-by-hop headers:
- `Connection`
- `Keep-Alive`
- `Proxy-Authenticate`
- `Proxy-Authorization`
- `Te`
- `Trailers`
- `Transfer-Encoding`
- `Upgrade`

### Response Headers Added

- `X-Cache` - Cache status (HIT, MISS, BYPASS)
- `X-Request-ID` - Unique request identifier

### Response Headers Forwarded

All upstream headers except hop-by-hop headers (same list as request headers).

---

## Examples

### Cache Hit Scenario

```bash
# First request (cache miss)
curl -i http://localhost:8080/api/data
# X-Cache: MISS

# Second request (cache hit)
curl -i http://localhost:8080/api/data
# X-Cache: HIT
```

### Non-Cacheable Request

```bash
# POST request bypasses cache
curl -X POST http://localhost:8080/api/data
# X-Cache: BYPASS

# Request with Authorization header
curl -H "Authorization: Bearer token" http://localhost:8080/api/data
# X-Cache: BYPASS
```

### Query Parameters

```bash
# Different query parameters = different cache entries
curl http://localhost:8080/api/users?page=1  # MISS → cached
curl http://localhost:8080/api/users?page=2  # MISS → cached (different key)
curl http://localhost:8080/api/users?page=1  # HIT (same key as first)
```

---

## Limitations

1. Only HTTP (not HTTPS) upstream connections currently supported
2. No support for WebSocket connections
3. No support for request/response body streaming (entire body buffered)
4. No support for cache invalidation (must wait for TTL expiry)
5. No support for conditional requests (If-None-Match, If-Modified-Since)
