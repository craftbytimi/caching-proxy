package proxy

import "testing"

func TestService_CacheHitMiss(t *testing.T) {
	// TODO: Implement integration test
	// 1. Create mock store
	// 2. Create test upstream
	// 3. Create proxy service
	// 4. First request should be MISS
	// 5. Second request should be HIT
	t.Skip("Integration test not implemented yet")
}

func TestService_NonCacheableRequest(t *testing.T) {
	// TODO: Test that POST requests bypass cache
	t.Skip("Non-cacheable request test not implemented yet")
}

func TestService_UpstreamError(t *testing.T) {
	// TODO: Test handling of upstream errors
	t.Skip("Upstream error test not implemented yet")
}
