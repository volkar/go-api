package cache

import (
	db "api/internal/platform/database/sqlc"
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var resolveAlbumBySlugsScript = redis.NewScript(`
    local user_slug_key = KEYS[1]
    local album_slug_prefix = ARGV[1]
    local album_entity_prefix = ARGV[2]
    local target_album_slug = ARGV[3]

    -- Get User ID
    local user_id = redis.call("GET", user_slug_key)
    if not user_id then
        return {err = "user_not_found"}
    end

    -- Get Album ID
    local album_slug_key = album_slug_prefix .. user_id .. "/" .. target_album_slug
    local album_id = redis.call("GET", album_slug_key)
    if not album_id then
        return {err = "album_not_found"}
    end

    -- Get Album
    local album_entity_key = album_entity_prefix .. album_id
    local album_json = redis.call("GET", album_entity_key)

    return album_json
`)

var invalidateUserWithSlugScript = redis.NewScript(`
	local user_entity_key = KEYS[1]
	local mapper_prefix = ARGV[1]

	local user_json = redis.call("GET", user_entity_key)
	if not user_json then
		return 0
	end

	local ok, user_obj = pcall(cjson.decode, user_json)

	if ok and type(user_obj) == "table" and user_obj.slug then
		local mapper_key = mapper_prefix .. user_obj.slug
		redis.call("UNLINK", mapper_key, user_entity_key)
		return 2
	end

	redis.call("UNLINK", user_entity_key)
	return 1
`)

/* Get album by slug from cache */
func (m *Manager) GetAlbumBySlugs(ctx context.Context, userSlug, albumSlug string) (db.Album, error) {
	userSlugKey := keyUserMapper(userSlug)

	// Execute the Lua script
	res, err := resolveAlbumBySlugsScript.Run(ctx, m.client,
		[]string{userSlugKey}, // KEYS
		AlbumMapperPrefix,     // ARGV[1]
		AlbumEntityPrefix,     // ARGV[2]
		albumSlug,             // ARGV[3]
	).Result()

	if res == nil {
		// Album not found in cache
		return db.Album{}, err
	}

	// Unmarshal album from cache
	var album db.Album
	if err := json.Unmarshal([]byte(res.(string)), &album); err == nil {
		return album, nil
	}

	return db.Album{}, err
}

/* Set album with mapper to cache */
func (m *Manager) SetAlbum(ctx context.Context, album db.Album) {
	// Set album to cache
	if err := m.set(ctx, keyAlbumEntity(album.ID), album); err != nil {
		m.logger.Warn("Album entity cache update failed", "err", err)
	}
	// Set mapper (user_id/album_slug -> album_id) to cache
	if err := m.set(ctx, keyAlbumMapper(album.UserID, album.Slug), album.ID.String()); err != nil {
		m.logger.Warn("Album mapper cache update failed", "err", err)
	}
}

/* Get multiple albums from cache */
func (m *Manager) GetAlbumList(ctx context.Context, keys []uuid.UUID) (map[uuid.UUID]db.Album, error) {
	var cacheKeys []string
	for _, key := range keys {
		cacheKeys = append(cacheKeys, keyAlbumEntity(key))
	}

	albumsPtrs, err := MGetItems[db.Album](ctx, m.client, cacheKeys)
	if err != nil {
		return nil, err
	}

	albumsArray := make(map[uuid.UUID]db.Album, len(keys))
	for i, albumPtr := range albumsPtrs {
		if albumPtr != nil {
			albumsArray[keys[i]] = *albumPtr
		}
	}

	return albumsArray, err
}

/* Get user by slug from cache */
func (m *Manager) GetUserBySlug(ctx context.Context, userSlug string) (db.User, error) {
	var mapper string
	err := m.get(ctx, keyUserMapper(userSlug), &mapper)
	if err != nil {
		return db.User{}, err
	}
	// Mapper found, try to parse ID
	id, err := uuid.Parse(mapper)
	if err != nil {
		return db.User{}, err
	}
	return m.GetUser(ctx, id)
}

/* Get user from cache */
func (m *Manager) GetUser(ctx context.Context, userID uuid.UUID) (db.User, error) {
	var u db.User
	err := m.get(ctx, keyUserEntity(userID), &u)
	return u, err
}

/* Set user with mapper to cache */
func (m *Manager) SetUser(ctx context.Context, user db.User) {
	// Set user entity to cache
	if err := m.set(ctx, keyUserEntity(user.ID), user); err != nil {
		m.logger.Warn("User entity cache update failed", "err", err)
	}
	// Set mapper (slug -> id) to cache
	if err := m.set(ctx, keyUserMapper(user.Slug), user.ID.String()); err != nil {
		m.logger.Warn("User mapper cache update failed", "err", err)
	}
}

/* Invalidate user cache */
func (m *Manager) InvalidateUser(ctx context.Context, userID uuid.UUID) error {
	_, err := invalidateUserWithSlugScript.Run(ctx, m.client, []string{keyUserEntity(userID)}, UserMapperPrefix).Result()

	if err != nil && !errors.Is(err, redis.Nil) {
		m.logger.Warn("Failed to run invalidate user lua script", "err", err)
		return err
	}

	return nil
}
