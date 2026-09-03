package ratelimiter

import "fmt"

type RateLimitResult struct {
	Allowed      bool
	Remaining    int
	RetryAfterMs *int64
}

type Limiter interface {
	Allow(key string) RateLimitResult
}

type LimiterFactory struct{}

func (f *LimiterFactory) Create(config map[string]interface{}) Limiter {
	algo, _ := config["algorithm"].(string)
	cfg, _ := config["algoConfig"].(map[string]interface{})
	if cfg == nil {
		cfg = map[string]interface{}{}
	}

	switch algo {
	case "TokenBucket":
		capacity := cfg["capacity"].(int)
		refillRate := cfg["refillRatePerSecond"].(int)
		return NewTokenBucketLimiter(capacity, refillRate)
	case "SlidingWindow":
		maxRequests := cfg["maxRequests"].(int)
		windowMs := cfg["windowMs"].(int)
		return NewSlidingWindowLogLimiter(maxRequests, int64(windowMs))
	default:
		panic(fmt.Sprintf("unknown algo %s", algo))
	}
}
