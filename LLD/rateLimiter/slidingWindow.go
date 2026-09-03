package ratelimiter

import "time"

type requestLog struct {
	timestamps []int64
}

type SlidingWindowLogLimiter struct {
	maxRequests int
	windowMs    int64
	logs        map[string]*requestLog
}

func NewSlidingWindowLogLimiter(maxReq int, windowMs int64) *SlidingWindowLogLimiter {
	return &SlidingWindowLogLimiter{
		maxRequests: maxReq,
		windowMs:    windowMs,
		logs:        make(map[string]*requestLog),
	}
}

func (sl *SlidingWindowLogLimiter) Allow(key string) RateLimitResult {
	log := sl.getOrCreateLog(key)

	now := time.Now().UnixMilli()
	cutoff := now - sl.windowMs

	idx := 0
	for idx < len(log.timestamps) && log.timestamps[idx] < cutoff {
		idx++
	}

	log.timestamps = log.timestamps[idx:]

	if len(log.timestamps) < sl.maxRequests {
		log.timestamps = append(log.timestamps, now)
		remaining := sl.maxRequests - len(log.timestamps)
		return RateLimitResult{Allowed: true, Remaining: remaining}
	}

	oldest := log.timestamps[0]
	retryAfter := (oldest + sl.windowMs) - now
	return RateLimitResult{Allowed: false, Remaining: 0, RetryAfterMs: &retryAfter}
}

func (sl *SlidingWindowLogLimiter) getOrCreateLog(key string) *requestLog {
	log, ok := sl.logs[key]
	if ok {
		return log
	}

	log = &requestLog{
		timestamps: []int64{},
	}

	sl.logs[key] = log
	return log
}
