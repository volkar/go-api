package cache

import (
	db "api/internal/platform/database/sqlc"
	"context"

	"github.com/google/uuid"
)

/* User updated event */
func (m *Manager) OnUserUpdated(ctx context.Context, user db.User, oldUserSlug string) {
	// User slug changed. Invalidate user mapper cache
	if oldUserSlug != user.Slug {
		m.client.Unlink(ctx, keyUserMapper(oldUserSlug))
	}
	// Set new user cache
	m.SetUser(ctx, user)
}

/* User deleted event */
func (m *Manager) OnUserDeleted(ctx context.Context, userID uuid.UUID) {
	// Invalidate user cache
	m.InvalidateUser(ctx, userID)
}

/* User purged event */
func (m *Manager) OnUserPurged(ctx context.Context, userID uuid.UUID) {
	// Invalidate user cache
	m.InvalidateUser(ctx, userID)
}

/* Album created event */
func (m *Manager) OnAlbumCreated(ctx context.Context, album db.Album) {
	// Set album cache
	m.SetAlbum(ctx, album)
}

/* Album updated event */
func (m *Manager) OnAlbumUpdated(ctx context.Context, album db.Album, oldSlug string) {
	// Invalidate album mapper cache (album slug changed)
	if oldSlug != album.Slug {
		m.client.Unlink(ctx, keyAlbumMapper(album.UserID, oldSlug))
	}
	// Set new album cache
	m.SetAlbum(ctx, album)
}

/* Delete album event */
func (m *Manager) OnAlbumDeleted(ctx context.Context, album db.Album) {
	// Delete album keys (entity and mapper)
	m.client.Unlink(ctx, keyAlbumEntity(album.ID), keyAlbumMapper(album.UserID, album.Slug))
}
