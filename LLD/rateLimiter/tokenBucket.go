package ratelimiter

import (
	"math"
	"time"
)

type Bucket struct {
	tokens         float64
	lastRefillTime int64
}

type TokenBucketLimiter struct {
	capacity            int
	refillRatePerSecond int
	buckets             map[string]*Bucket
}

func NewTokenBucketLimiter(capacity, refillRatePerSecond int) *TokenBucketLimiter {
	return &TokenBucketLimiter{
		capacity:            capacity,
		refillRatePerSecond: refillRatePerSecond,
		buckets:             make(map[string]*Bucket),
	}
}

func (t *TokenBucketLimiter) Allow(key string) RateLimitResult {
	bucket := t.getOrCreateBucket(key)

	now := time.Now().UnixMilli()
	elapsed := now - bucket.lastRefillTime
	tokensToAdd := (float64(elapsed) * float64(t.refillRatePerSecond)) / 1000
	bucket.tokens = math.Min(float64(t.capacity), tokensToAdd+bucket.tokens)
	bucket.lastRefillTime = now

	if bucket.tokens >= 1 {
		bucket.tokens -= 1
		remaining := int(math.Floor(bucket.tokens))
		return RateLimitResult{Allowed: true, Remaining: remaining}
	}

	tokensNeeded := 1 - bucket.tokens
	retryAfter := int64(math.Ceil((tokensNeeded * 1000) / float64(t.refillRatePerSecond)))
	return RateLimitResult{
		Allowed:      false,
		Remaining:    0,
		RetryAfterMs: &retryAfter,
	}
}

func (t *TokenBucketLimiter) getOrCreateBucket(key string) *Bucket {
	bucket, exists := t.buckets[key]
	if exists {
		return bucket
	}
	return &Bucket{
		tokens:         float64(t.capacity),
		lastRefillTime: time.Now().UnixMilli(),
	}
}
