package ratelimit

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Single-call LUA scrip
var script = redis.NewScript(`
    local key = KEYS[1]
    local limit = tonumber(ARGV[1])
    local window = tonumber(ARGV[2])

    local current = redis.call("GET", key)
    if current and tonumber(current) >= limit then
        return {0, 0, redis.call("PTTL", key)}
    end

    local new_val = redis.call("INCR", key)
    if new_val == 1 then
        redis.call("PEXPIRE", key, window * 1000)
    end

    return {1, limit - new_val, redis.call("PTTL", key)}
`)

type Limiter struct {
	client *redis.Client
}

type RateLimitResult struct {
	Allowed   bool
	Remaining int
	Reset     time.Duration
}

func New(client *redis.Client) *Limiter {
	return &Limiter{client: client}
}

func (l *Limiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (RateLimitResult, error) {
	res, err := script.Run(ctx, l.client, []string{key}, limit, int(window.Seconds())).Slice()
	if err != nil {
		return RateLimitResult{}, err
	}

	allowed := res[0].(int64) == 1
	remaining := int(res[1].(int64))
	pttl := max(res[2].(int64), 0)
	reset := time.Duration(pttl) * time.Millisecond

	return RateLimitResult{
		Allowed:   allowed,
		Remaining: remaining,
		Reset:     reset,
	}, nil
}
