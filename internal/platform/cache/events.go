package cache

import (
	"context"

	"github.com/google/uuid"
)

/* User updated event */
func (m *Manager) OnUserUpdated(ctx context.Context, userID uuid.UUID, userSlug string, oldUserSlug string) {
	// User changed. Invalidate user cache
	m.invalidateUser(ctx, userID)

	if userSlug != oldUserSlug {
		// Slug changed, invalidate all albums cache
		m.invalidateAllAlbums(ctx, userID)
	}
}

/* User deleted event */
func (m *Manager) OnUserDeleted(ctx context.Context, userID uuid.UUID) {
	// Invalidate user cache
	m.invalidateUser(ctx, userID)
	// Invalidate album list cache
	m.invalidateAlbumList(ctx, userID)
	// Invalidate all albums cache
	m.invalidateAllAlbums(ctx, userID)
	// Invalidate deleted album list cache
	m.invalidateDeletedAlbumList(ctx, userID)
}

/* User purged event */
func (m *Manager) OnUserPurged(ctx context.Context, userID uuid.UUID) {
	// Invalidate user cache
	m.invalidateUser(ctx, userID)
	// Invalidate album list cache
	m.invalidateAlbumList(ctx, userID)
	// Invalidate all albums cache
	m.invalidateAllAlbums(ctx, userID)
	// Invalidate deleted album list cache
	m.invalidateDeletedAlbumList(ctx, userID)
}

/* Album created event */
func (m *Manager) OnAlbumCreated(ctx context.Context, userID uuid.UUID) {
	// Invalidate album list cache
	m.invalidateAlbumList(ctx, userID)
}

/* Album updated event */
func (m *Manager) OnAlbumUpdated(ctx context.Context, userID uuid.UUID, userSlug string, oldAlbumSlug string) {
	// Invalidate album list cache
	m.invalidateAlbumList(ctx, userID)
	// Invalidate album cache (album changed)
	m.invalidateAlbum(ctx, userSlug, oldAlbumSlug)
}

/* Delete album event */
func (m *Manager) OnAlbumDeleted(ctx context.Context, userID uuid.UUID, userSlug string, oldAlbumSlug string) {
	// Invalidate album list cache
	m.invalidateAlbumList(ctx, userID)
	// Invalidate deleted album list cache
	m.invalidateDeletedAlbumList(ctx, userID)
	// Invalidate album cache (album deleted)
	m.invalidateAlbum(ctx, userSlug, oldAlbumSlug)
}

/* Delete all albums event */
func (m *Manager) OnAllAlbumsDeleted(ctx context.Context, userID uuid.UUID) {
	// Invalidate user cache (album list changed)
	m.invalidateUser(ctx, userID)
	// Invalidate album list cache
	m.invalidateAlbumList(ctx, userID)
	// Invalidate deleted album list cache
	m.invalidateDeletedAlbumList(ctx, userID)
	// Invalidate all albums cache
	m.invalidateAllAlbums(ctx, userID)
}

/* Album restored event */
func (m *Manager) OnAlbumRestored(ctx context.Context, userID uuid.UUID) {
	// Invalidate album list cache
	m.invalidateAlbumList(ctx, userID)
	// Invalidate deleted album list cache
	m.invalidateDeletedAlbumList(ctx, userID)
}

/* Album purged event */
func (m *Manager) OnAlbumPurged(ctx context.Context, userID uuid.UUID) {
	// Invalidate deleted album list cache
	m.invalidateDeletedAlbumList(ctx, userID)
}
