package core

import (
	"context"
	"testing"
)

func TestSessionKeyContextRoundTrip(t *testing.T) {
	ctx := context.Background()
	if got := SessionKeyFromContext(ctx); got != "" {
		t.Fatalf("empty ctx: want \"\", got %q", got)
	}

	key := "feishu:oc_abc:om_123"
	ctx = WithSessionKey(ctx, key)
	if got := SessionKeyFromContext(ctx); got != key {
		t.Fatalf("round-trip: want %q, got %q", key, got)
	}

	// Empty key must not wrap (so it can't shadow a real key set upstream).
	if WithSessionKey(context.Background(), "") != context.Background() {
		t.Fatal("WithSessionKey with empty key should return ctx unchanged")
	}
}
