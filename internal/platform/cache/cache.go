package cache

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type Manager struct {
	client *redis.Client
	ttl    time.Duration
}

func New(client *redis.Client, defaultTTL time.Duration) *Manager {
	return &Manager{
		client: client,
		ttl:    defaultTTL,
	}
}

/* Get redis client */
func (m *Manager) Client() *redis.Client {
	return m.client
}

/* Get from cache with cast to target */
func (m *Manager) Get(ctx context.Context, key string, target any) error {
	val, err := m.client.Get(ctx, key).Result()
	if err == nil {
		// If value is string or int, set it directly. Otherwise unmarshal it from JSON.
		if strVal, ok := target.(*string); ok {
			*strVal = val
			return nil
		}
		if intVal, ok := target.(*int); ok {
			*intVal, err = strconv.Atoi(val)
			return err
		}
		return json.Unmarshal([]byte(val), target)
	}
	return err
}

/* Set to cache */
func (m *Manager) Set(ctx context.Context, key string, data any) error {
	var d []byte
	var err error
	// If data is string or int, set it directly. Otherwise marshal it to JSON.
	if strVal, ok := data.(string); ok {
		d = []byte(strVal)
	} else if intVal, ok := data.(int); ok {
		d = []byte(strconv.Itoa(intVal))
	} else {
		d, err = json.Marshal(data)
		if err != nil {
			return err
		}
	}
	return m.client.Set(ctx, key, d, m.ttl).Err()
}

/* Delete keys from cache */
func (m *Manager) Unlink(ctx context.Context, keys ...string) error {
	return m.client.Unlink(ctx, keys...).Err()
}

/* Run LUA script */
func (m *Manager) RunScript(ctx context.Context, script *redis.Script, keys []string, args ...any) (any, error) {
	return script.Run(ctx, m.client, keys, args...).Result()
}

/* Get multiple items from cache with type T */
func MGetItems[T any](ctx context.Context, client redis.Cmdable, keys []string) ([]*T, error) {
	if len(keys) == 0 {
		return nil, nil
	}

	vals, err := client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	result := make([]*T, len(vals))

	for i, val := range vals {
		if val == nil {
			continue
		}
		strVal, ok := val.(string)
		if !ok {
			continue
		}
		var item T
		if err := json.Unmarshal([]byte(strVal), &item); err == nil {
			result[i] = &item
		}
	}
	return result, nil
}

/* Clears all cached data. For development purposes only */
func (m *Manager) ClearFullCache(ctx context.Context) error {
	return m.client.FlushDBAsync(ctx).Err()
}
