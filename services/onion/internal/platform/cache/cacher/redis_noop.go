package cacher

import (
	"context"
	"time"
)

// noopRedisCache is returned when Redis is disabled; all operations are silent no-ops.
type noopRedisCache struct{}

func (n *noopRedisCache) Put(_ context.Context, _ string, _ []byte, _ time.Duration) {}
func (n *noopRedisCache) Get(_ context.Context, _ string) []byte                     { return nil }
func (n *noopRedisCache) GetDel(_ context.Context, _ string) []byte                  { return nil }
func (n *noopRedisCache) PutMulti(_ context.Context, _ map[string][]byte, _ time.Duration) {
}

func (n *noopRedisCache) GetMulti(_ context.Context, keys []string) (map[string][]byte, []string) {
	return make(map[string][]byte), keys
}
func (n *noopRedisCache) Delete(_ context.Context, _ []string) {}
func (n *noopRedisCache) DeleteByPattern(ctx context.Context, pattern string) int {
	return GetInMemoryCache(ctx).DeleteByPattern(ctx, pattern)
}
func (n *noopRedisCache) HSet(_ context.Context, _, _ string, _ []byte, _ time.Duration) {}
func (n *noopRedisCache) HFieldExpire(_ context.Context, _, _ string, _ time.Duration)   {}
func (n *noopRedisCache) HSetWithFieldExpiry(_ context.Context, _, _ string, _ []byte, _, _ time.Duration) {
}
func (n *noopRedisCache) HDel(_ context.Context, _, _ string) bool   { return true }
func (n *noopRedisCache) HGet(_ context.Context, _, _ string) []byte { return nil }
func (n *noopRedisCache) HGetAll(_ context.Context, _ string) map[string][]byte {
	return make(map[string][]byte)
}

func (n *noopRedisCache) ZAdd(_ context.Context, _ string, _ []byte, _ float64, _ time.Duration) bool {
	return true
}
func (n *noopRedisCache) ZRem(_ context.Context, _ string, _ []byte) bool                { return true }
func (n *noopRedisCache) ZRemRangeUpToScore(_ context.Context, _ string, _ float64) bool { return true }
func (n *noopRedisCache) ZRemRangeByRank(_ context.Context, _ string, _, _ int64) bool   { return true }
func (n *noopRedisCache) ZRange(_ context.Context, _ string, _, _ int64) [][]byte        { return nil }
func (n *noopRedisCache) ZCard(_ context.Context, _ string) (int64, bool)                { return 0, true }
func (n *noopRedisCache) SAdd(_ context.Context, _ string, _ []byte, _ time.Duration) bool {
	return true
}
func (n *noopRedisCache) SRem(_ context.Context, _ string, _ []byte) bool                  { return true }
func (n *noopRedisCache) SMembers(_ context.Context, _ string) [][]byte                    { return nil }
func (n *noopRedisCache) SCard(_ context.Context, _ string) (int64, bool)                  { return 0, true }
func (n *noopRedisCache) HSetNx(_ context.Context, _, _ string, _ []byte, _ time.Duration) {}
func (n *noopRedisCache) Increment(_ context.Context, _ string, _ time.Duration) int64     { return 0 }
func (n *noopRedisCache) GetInt(_ context.Context, _ string) []byte                        { return nil }
func (n *noopRedisCache) SetNx(_ context.Context, _ string, _ []byte, _ time.Duration) bool {
	return true
}
func (n *noopRedisCache) Expire(_ context.Context, _ string, _ time.Duration)   {}
func (n *noopRedisCache) ExpireAt(_ context.Context, _ string, _ time.Duration) {}

// PutConfirmed never persists anything, so it must report failure rather than
// let callers treat a no-op as a successful write.
func (n *noopRedisCache) PutConfirmed(_ context.Context, _ string, _ []byte, _ time.Duration) bool {
	return false
}
