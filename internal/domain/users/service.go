package users

import (
	"api/internal/domain/albums"
	"api/internal/domain/tokens"
	"api/internal/platform/response"
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Service struct {
	users  *Repository
	albums AlbumLister
	tokens *tokens.Manager
}

func NewService(repo *Repository, albums AlbumLister, tokens *tokens.Manager) *Service {
	return &Service{
		users:  repo,
		albums: albums,
		tokens: tokens,
	}
}

type AlbumLister interface {
	ListAvailable(ctx context.Context, userID uuid.UUID, viewerID uuid.UUID, viewerEmail string, cursor string, limit int) ([]albums.AlbumInList, string, error)
}

/* Get non deleted user info by id */
func (s *Service) GetAvailable(ctx context.Context, userID uuid.UUID) (User, error) {
	u, err := s.users.GetAvailable(ctx, userID)
	if err != nil {
		return User{}, response.ErrUserNotFound.Wrap(err)
	}
	return u, nil
}

/* Get non deleted user info by slug */
func (s *Service) GetAvailableBySlug(ctx context.Context, userSlug string) (User, error) {
	u, err := s.users.GetAvailableBySlug(ctx, userSlug)
	if err != nil {
		return User{}, response.ErrUserNotFound.Wrap(err)
	}
	return u, nil
}

/* Get non deleted album list by slug */
func (s *Service) AlbumList(ctx context.Context, userSlug string, viewerID uuid.UUID, viewerEmail string, cursor string, limit int) ([]albums.AlbumInList, string, error) {
	// Get user
	u, err := s.users.GetAvailableBySlug(ctx, userSlug)
	if err != nil {
		// User not found
		if errors.Is(err, pgx.ErrNoRows) {
			return []albums.AlbumInList{}, "", response.ErrUserNotFound.Wrap(err)
		}
		return []albums.AlbumInList{}, "", err
	}

	// Get albums
	a, nextCursor, err := s.albums.ListAvailable(ctx, u.ID, viewerID, viewerEmail, cursor, limit)
	if err != nil {
		// List error
		return []albums.AlbumInList{}, "", response.ErrUserNotFound.Wrap(err)
	}

	return a, nextCursor, nil
}

/* Upsert confirmed user */
func (s *Service) Upsert(ctx context.Context, email string, username string) (User, error) {
	return s.users.Upsert(ctx, email, username)
}

/* Create user. Use with caution! Users must be created with Upsert function via OAuth process and have validated email */
func (s *Service) Create(ctx context.Context, email string, username string, slug string, role Role) (User, error) {
	return s.users.Create(ctx, email, username, slug, string(role))
}

/* Update user info */
func (s *Service) Update(ctx context.Context, actorID uuid.UUID, targetUserID uuid.UUID, userSlug string, username string) (User, error) {
	// Check if user have permission
	if actorID != targetUserID {
		return User{}, response.ErrNoPermission
	}
	u, err := s.users.Update(ctx, targetUserID, username, userSlug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// User not found
			return User{}, response.ErrUserNotFound.Wrap(err)
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if (pgErr.Code == "23505") && (pgErr.ConstraintName == "users_slug_key") {
				// Slug already exists
				return User{}, response.ErrUserSlugExists.Wrap(err)
			}
		}
		return User{}, err
	}
	return u, nil
}

/* Delete user */
func (s *Service) Delete(ctx context.Context, actorID uuid.UUID, targetUserID uuid.UUID) (uuid.UUID, error) {
	// Check if user have permission
	if actorID != targetUserID {
		return uuid.Nil, response.ErrNoPermission
	}

	// Delete user
	id, err := s.users.Delete(ctx, targetUserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, response.ErrNoPermission.Wrap(err)
		}
		return uuid.Nil, err
	}

	// Delete all user tokens
	s.tokens.DeleteAllRefreshForUser(ctx, targetUserID)

	return id, err
}
