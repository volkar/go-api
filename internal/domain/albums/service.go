package albums

import (
	"api/internal/domain/albums/albumtypes"
	"api/internal/platform/response"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Service struct {
	albums *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{
		albums: repo,
	}
}

/* Get available album by user slug and album slug */
func (s *Service) GetAvailable(ctx context.Context, userSlug string, albumSlug string, viewerID uuid.UUID, viewerEmail string) (Album, error) {
	a, err := s.albums.GetAvailable(ctx, userSlug, albumSlug, viewerID, viewerEmail)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Album{}, response.ErrAlbumNotFound.Wrap(err)
		}
		return Album{}, err
	}
	// Album found in cache or DB
	// If it comes from cache, we need to check access permissions and if album is active
	isOwner := viewerID == a.UserID
	if albumtypes.CanAccess(a.Access, a.SharedEmails, viewerEmail, isOwner) == false || a.IsActive == false {
		return Album{}, response.ErrAlbumNotFound
	}
	return a, nil
}

func (s *Service) ListAvailable(ctx context.Context, userID uuid.UUID, viewerID uuid.UUID, viewerEmail string, cursor string, limit int) ([]AlbumInList, string, error) {
	if limit <= 0 || limit > 60 {
		limit = 60
	}

	a, nextCursor, err := s.albums.ListAvailable(ctx, userID, viewerID, viewerEmail, cursor, int32(limit))
	if err != nil {
		return []AlbumInList{}, "", err
	}
	// Map Albums to AlbumInList
	albums := ToAlbumList(a)
	return albums, nextCursor, nil
}

/* Get list of deleted albums by user id */
func (s *Service) ListDeleted(ctx context.Context, userID uuid.UUID, cursor string, limit int) ([]AlbumInList, string, error) {
	if limit <= 0 || limit > 60 {
		limit = 60
	}
	a, nextCursor, err := s.albums.ListDeleted(ctx, userID, cursor, int32(limit))
	if err != nil {
		return []AlbumInList{}, "", err
	}
	// Map Albums to AlbumInList
	albums := ToAlbumList(a)
	return albums, nextCursor, nil
}

/* Create album */
func (s *Service) Create(ctx context.Context, userID uuid.UUID, title string, slug string, atlas albumtypes.Atlas, access string, share []string, dateAt time.Time) (Album, error) {
	a, err := s.albums.Create(ctx, title, slug, atlas, access, share, dateAt, userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if (pgErr.Code == "23505") && (pgErr.ConstraintName == "idx_albums_user_slug_active") {
				return Album{}, response.ErrAlbumSlugExists.Wrap(err)
			}
		}
		if errors.Is(err, pgx.ErrNoRows) {
			// User deleted or not existed
			return Album{}, response.ErrNoPermission.Wrap(err)
		}
		return Album{}, err
	}
	return a, nil
}

/* Update album */
func (s *Service) Update(ctx context.Context, userID uuid.UUID, albumID uuid.UUID, title string, slug string, atlas albumtypes.Atlas, access string, sharedEmails []string, dateAt time.Time, isActive bool) (Album, error) {
	a, err := s.albums.Update(ctx, userID, albumID, title, slug, atlas, access, sharedEmails, dateAt, isActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Album not found or user is deleted
			return Album{}, response.ErrNoPermission.Wrap(err)
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if (pgErr.Code == "23505") && (pgErr.ConstraintName == "idx_albums_user_slug_active") {
				// Slug conflict
				return Album{}, response.ErrAlbumSlugExists.Wrap(err)
			}
		}
		return Album{}, err
	}
	return a, nil
}

/* Delete album */
func (s *Service) Delete(ctx context.Context, userID uuid.UUID, albumID uuid.UUID) (uuid.UUID, error) {
	a, err := s.albums.Delete(ctx, userID, albumID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Album not found or user is deleted
			return uuid.Nil, response.ErrNoPermission.Wrap(err)
		}
		return uuid.Nil, err
	}
	return a, nil
}

/* Restore deleted album */
func (s *Service) Restore(ctx context.Context, userID uuid.UUID, albumID uuid.UUID) (uuid.UUID, error) {
	a, err := s.albums.Restore(ctx, userID, albumID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// User is deleted or album not found
			return uuid.Nil, response.ErrNoPermission.Wrap(err)
		}
		return uuid.Nil, err
	}
	return a, nil
}

/* Purge deleted album */
func (s *Service) Purge(ctx context.Context, userID uuid.UUID, albumID uuid.UUID) (uuid.UUID, error) {
	a, err := s.albums.Purge(ctx, userID, albumID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// User is deleted or album not found
			return uuid.Nil, response.ErrNoPermission.Wrap(err)
		}
		return uuid.Nil, err
	}
	return a, nil
}
