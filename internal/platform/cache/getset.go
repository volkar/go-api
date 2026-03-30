package cache

import (
	"api/internal/domain/albums"
	"api/internal/domain/users"
	"context"

	"github.com/google/uuid"
)

/* Get album from cache */
func (m *Manager) GetAlbum(ctx context.Context, userSlug string, albumSlug string) (albums.Album, error) {
	var a albums.Album
	err := m.get(ctx, keyAlbum(userSlug, albumSlug), &a)
	return a, err
}

/* Set album to cache */
func (m *Manager) SetAlbum(ctx context.Context, userSlug string, albumSlug string, album albums.Album) {
	if err := m.setWithTag(ctx, keyAlbum(userSlug, albumSlug), "owner_id:"+album.UserID.String(), album); err != nil {
		m.logger.Warn("cache update failed", "err", err)
	}
}

/* Get available albums from cache */
func (m *Manager) GetAvailableAlbums(ctx context.Context, userID uuid.UUID) ([]albums.AlbumInList, error) {
	var a []albums.AlbumInList
	err := m.get(ctx, keyAlbumList(userID), &a)
	return a, err
}

/* Set available albums to cache */
func (m *Manager) SetAvailableAlbums(ctx context.Context, userID uuid.UUID, albums []albums.AlbumInList) {
	if err := m.set(ctx, keyAlbumList(userID), albums); err != nil {
		m.logger.Warn("cache update failed", "err", err)
	}
}

/* Get deleted albums from cache */
func (m *Manager) GetDeletedAlbums(ctx context.Context, userID uuid.UUID) ([]albums.AlbumInList, error) {
	var a []albums.AlbumInList
	err := m.get(ctx, keyDeletedAlbums(userID), &a)
	return a, err
}

/* Set album list to cache */
func (m *Manager) SetDeletedAlbums(ctx context.Context, userID uuid.UUID, albums []albums.AlbumInList) {
	if err := m.set(ctx, keyDeletedAlbums(userID), albums); err != nil {
		m.logger.Warn("cache update failed", "err", err)
	}
}

/* Get user by slug from cache */
func (m *Manager) GetUserBySlug(ctx context.Context, userSlug string) (users.User, error) {
	var mapper UserMapper
	err := m.get(ctx, keyMapper(userSlug), &mapper)
	if err != nil {
		return users.User{}, err
	}
	// Mapper found, try to parse ID
	id, err := uuid.Parse(mapper.ID)
	if err != nil {
		return users.User{}, err
	}
	return m.GetUser(ctx, id)
}

/* Get user from cache */
func (m *Manager) GetUser(ctx context.Context, userID uuid.UUID) (users.User, error) {
	var u users.User
	err := m.get(ctx, keyUser(userID), &u)
	return u, err
}

/* Set user with mapper to cache */
func (m *Manager) SetUser(ctx context.Context, user users.User) {
	// Set user to cache
	if err := m.setWithTag(ctx, keyUser(user.ID), "user_id:"+user.ID.String(), user); err != nil {
		m.logger.Warn("cache update failed", "err", err)
	}
	// Set mapper (slug->id) to cache
	mapper := UserMapper{
		ID: user.ID.String(),
	}
	if err := m.setWithTag(ctx, keyMapper(user.Slug), "user_id:"+user.ID.String(), mapper); err != nil {
		m.logger.Warn("cache update failed", "err", err)
	}
}
