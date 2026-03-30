package cache

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Manager struct {
	client   *redis.Client
	logger   *slog.Logger
	ttl      time.Duration
	graceTTL time.Duration
}

type UserMapper struct {
	ID string `json:"id"`
}

func New(client *redis.Client, defaultTTL time.Duration, logger *slog.Logger) *Manager {
	return &Manager{
		client:   client,
		logger:   logger,
		ttl:      defaultTTL,
		graceTTL: 5 * time.Minute,
	}
}

/* Get from cache with cast to target */
func (m *Manager) get(ctx context.Context, key string, target any) error {
	val, err := m.client.Get(ctx, key).Result()
	if err == nil {
		return json.Unmarshal([]byte(val), target)
	}
	return err
}

/* Set to cache */
func (m *Manager) set(ctx context.Context, key string, data any) error {
	// Detached context with timeout
	detachedCtx := context.WithoutCancel(ctx)
	bgCtx, cancel := context.WithTimeout(detachedCtx, 100*time.Millisecond)
	defer cancel()

	d, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return m.client.Set(bgCtx, key, d, m.ttl).Err()
}

/* Set to cache with tag */
func (m *Manager) setWithTag(ctx context.Context, key string, tag string, data any) error {
	// Detached context with timeout
	detachedCtx := context.WithoutCancel(ctx)
	bgCtx, cancel := context.WithTimeout(detachedCtx, 100*time.Millisecond)
	defer cancel()

	d, err := json.Marshal(data)
	if err != nil {
		return err
	}

	pipe := m.client.Pipeline()
	pipe.Set(bgCtx, key, d, m.ttl)
	pipe.SAdd(bgCtx, tag, key)
	pipe.Expire(bgCtx, tag, m.ttl+m.graceTTL)
	_, err = pipe.Exec(bgCtx)

	return err
}

/* Invalidates user cache with mapper by user id */
func (m *Manager) invalidateUser(ctx context.Context, tag uuid.UUID) error {
	return m.invalidateTag(ctx, "user_id:"+tag.String())
}

/* Invalidates album cache by user slug and album slug */
func (m *Manager) invalidateAlbum(ctx context.Context, userSlug string, albumSlug string) error {
	return m.client.Del(ctx, keyAlbum(userSlug, albumSlug)).Err()
}

/* Invalidates album list cache by user id */
func (m *Manager) invalidateAlbumList(ctx context.Context, userID uuid.UUID) error {
	return m.client.Del(ctx, keyAlbumList(userID)).Err()
}

/* Invalidates deleted album list cache by user id */
func (m *Manager) invalidateDeletedAlbumList(ctx context.Context, userID uuid.UUID) error {
	return m.client.Del(ctx, keyDeletedAlbums(userID)).Err()
}

/* Invalidates all albums cache by user id */
func (m *Manager) invalidateAllAlbums(ctx context.Context, tag uuid.UUID) error {
	tagKey := "owner_id:" + tag.String()
	return m.invalidateTag(ctx, tagKey)
}

/* Delete all keys with tag */
func (m *Manager) invalidateTag(ctx context.Context, tagKey string) error {
	// Get all keys with tag
	keys, err := m.client.SMembers(ctx, tagKey).Result()
	if err != nil {
		return err
	}

	if len(keys) == 0 {
		return nil
	}

	// Add tag key to keys
	keys = append(keys, tagKey)
	// Delete all keys
	return m.client.Unlink(ctx, keys...).Err()
}

/* Clears all cache. For development purposes only */
func (m *Manager) ClearFullCache(ctx context.Context) error {
	return m.client.FlushDBAsync(ctx).Err()
}
