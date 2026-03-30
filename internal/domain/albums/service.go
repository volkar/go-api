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
	a, err := s.albums.GetAvailable(ctx, userSlug, albumSlug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Album{}, response.ErrAlbumNotFound.Wrap(err)
		}
		return Album{}, err
	}
	// Check album access
	isOwner := a.UserID == viewerID
	if !albumtypes.CanAccess(a.Access, viewerEmail, isOwner) {
		return Album{}, response.ErrAlbumNotFound
	}
	return a, nil
}

/* Get list of available albums by user id */
func (s *Service) ListAvailable(ctx context.Context, userID uuid.UUID, viewerEmail string, isOwner bool) ([]AlbumInList, error) {
	a, err := s.albums.ListAvailable(ctx, userID)
	if err != nil {
		return []AlbumInList{}, err
	}
	// Filter albums by access property
	filteredAlbums := []AlbumInList{}
	for i := range a {
		if albumtypes.CanAccess(a[i].Access, viewerEmail, isOwner) {
			filteredAlbums = append(filteredAlbums, a[i])
		}
	}
	return filteredAlbums, nil
}

/* Get list of deleted albums by user id */
func (s *Service) ListDeleted(ctx context.Context, userID uuid.UUID) ([]AlbumInList, error) {
	a, err := s.albums.ListDeleted(ctx, userID)
	if err != nil {
		return []AlbumInList{}, response.ErrAlbumsNotFound.Wrap(err)
	}
	return a, nil
}

/* Create album */
func (s *Service) Create(ctx context.Context, userID uuid.UUID, title string, slug string, atlas albumtypes.Atlas, access albumtypes.Access, dateAt time.Time) (Album, error) {
	a, err := s.albums.Create(ctx, title, slug, atlas, access, dateAt, userID)
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
func (s *Service) Update(ctx context.Context, userID uuid.UUID, albumID uuid.UUID, title string, slug string, atlas albumtypes.Atlas, access albumtypes.Access, dateAt time.Time, isActive bool) (Album, error) {
	a, err := s.albums.Update(ctx, userID, albumID, title, slug, atlas, access, dateAt, isActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Album not found or user is deleted
			return Album{}, response.ErrNoPermission.Wrap(err)
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if (pgErr.Code == "23505") && (pgErr.ConstraintName == "albums_user_id_slug_key") {
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

/* Delete all albums */
func (s *Service) DeleteAll(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	a, err := s.albums.DeleteAll(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// User is deleted
			return []uuid.UUID{}, response.ErrNoPermission.Wrap(err)
		}
		return []uuid.UUID{}, err
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
