package ratelimiter

func NewRateLimiter(configs []map[string]interface{}, defaultConfig map[string]interface{}) *RateLimiter {
	factory := &LimiterFactory{}
	limiters := make(map[string]Limiter)

	for _, config := range configs {
		endpointValue, ok := config["endpoint"]
		if !ok {
			continue
		}
		endpoint, ok := endpointValue.(string)
		if !ok || endpoint == "" {
			continue
		}
		limiters[endpoint] = factory.Create(config)
	}

	return &RateLimiter{
		limiters:       limiters,
		defaultLimiter: factory.Create(defaultConfig),
	}
}

type RateLimiter struct {
	limiters       map[string]Limiter
	defaultLimiter Limiter
}

func (r *RateLimiter) Allow(clientID string, endpoint string) RateLimitResult {
	limiter, ok := r.limiters[endpoint]
	if !ok {
		limiter = r.defaultLimiter
	}
	return limiter.Allow(clientID)
}
