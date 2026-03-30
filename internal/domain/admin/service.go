package admin

import (
	"api/internal/platform/response"
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Service struct {
	admin *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{
		admin: repo,
	}
}

/* Hard delete user (with all albums via db onDelete) */
func (s *Service) PurgeUser(ctx context.Context, actorID uuid.UUID, actorRole string, targetID uuid.UUID) (uuid.UUID, error) {
	// Check if actor role is admin and admin is not deleting himself
	if actorRole != "admin" || actorID == targetID {
		return uuid.Nil, response.ErrNoPermission
	}

	uID, err := s.admin.PurgeUser(ctx, targetID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, response.ErrUserNotFound.Wrap(err)
		}
		return uuid.Nil, err
	}
	return uID, nil
}

/* Restore deleted user */
func (s *Service) RestoreUser(ctx context.Context, actorID uuid.UUID, actorRole string, targetUserID uuid.UUID) (uuid.UUID, string, error) {
	// Check if actor role is admin and admin is not restoring himself
	if actorRole != "admin" || actorID == targetUserID {
		return uuid.Nil, "", response.ErrNoPermission
	}
	id, slug, err := s.admin.RestoreUser(ctx, targetUserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, "", response.ErrUserNotFound.Wrap(err)
		}
		return uuid.Nil, "", err
	}
	return id, slug, nil
}
