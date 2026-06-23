package cacher

import (
	"context"
	"errors"
	"testing"
	"time"
)

func initTestCache(t *testing.T) context.Context {
	t.Helper()
	ctx := context.Background()
	if err := InitializeLocalCache(ctx); err != nil {
		t.Fatalf("InitializeLocalCache: %v", err)
	}
	return ctx
}

func TestGetOrComputeSingleUsesCachedValue(t *testing.T) {
	ctx := initTestCache(t)
	cache := OfWithNamespace[string]("test-single-hit")
	key := t.Name()
	t.Cleanup(func() { cache.Delete(ctx, []string{key}, MemoryOnly()) })

	fetches := 0
	fetch := func(context.Context) (string, error) {
		fetches++
		return "fresh", nil
	}

	first, err := cache.GetOrComputeSingle(ctx, key, fetch, MemoryOnly(), WithExpiry(ctx, time.Minute))
	if err != nil {
		t.Fatalf("first GetOrComputeSingle: %v", err)
	}
	second, err := cache.GetOrComputeSingle(ctx, key, fetch, MemoryOnly(), WithExpiry(ctx, time.Minute))
	if err != nil {
		t.Fatalf("second GetOrComputeSingle: %v", err)
	}

	if first != "fresh" || second != "fresh" {
		t.Fatalf("values = %q/%q, want fresh/fresh", first, second)
	}
	if fetches != 1 {
		t.Fatalf("fetches = %d, want 1", fetches)
	}
}

func TestGetOrComputeSingleCanForceOriginFetch(t *testing.T) {
	ctx := initTestCache(t)
	cache := OfWithNamespace[string]("test-single-force-origin")
	key := t.Name()
	t.Cleanup(func() { cache.Delete(ctx, []string{key}, MemoryOnly()) })

	cache.Put(ctx, key, "cached", MemoryOnly(), WithExpiry(ctx, time.Minute))
	got, err := cache.GetOrComputeSingle(ctx, key, func(context.Context) (string, error) {
		return "fresh", nil
	}, MemoryOnly(), WithRandomInvalidation(1), WithExpiry(ctx, time.Minute))
	if err != nil {
		t.Fatalf("GetOrComputeSingle: %v", err)
	}
	if got != "fresh" {
		t.Fatalf("GetOrComputeSingle = %q, want fresh", got)
	}
}

func TestGetOrComputeSinglePropagatesFetchError(t *testing.T) {
	ctx := initTestCache(t)
	cache := OfWithNamespace[string]("test-single-error")
	key := t.Name()
	wantErr := errors.New("origin failed")

	got, err := cache.GetOrComputeSingle(ctx, key, func(context.Context) (string, error) {
		return "", wantErr
	}, MemoryOnly(), WithExpiry(ctx, time.Minute))
	if !errors.Is(err, wantErr) {
		t.Fatalf("err = %v, want %v", err, wantErr)
	}
	if got != "" {
		t.Fatalf("value = %q, want empty", got)
	}
}

func TestRemoteOnlyBypassesLocalCache(t *testing.T) {
	ctx := initTestCache(t)
	cache := OfWithNamespace[string]("test-remote-only")
	key := t.Name()
	t.Cleanup(func() { cache.Delete(ctx, []string{key}, MemoryOnly()) })

	cache.Put(ctx, key, "local", MemoryOnly(), WithExpiry(ctx, time.Minute))
	if got, ok := cache.Get(ctx, key, RemoteOnly()); ok {
		t.Fatalf("RemoteOnly Get hit local cache: %q", got)
	}
}

func TestPatternKeysUseModelNamespace(t *testing.T) {
	cache := OfWithNamespace[string]("test-wildcard")

	got := cache.transformListKeys([]string{"purchase_orders:1:*"})
	if len(got) != 1 || got[0] != "test-wildcard-purchase_orders:1:*" {
		t.Fatalf("pattern key = %#v, want namespaced wildcard", got)
	}
}
