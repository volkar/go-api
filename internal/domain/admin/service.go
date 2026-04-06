package admin

import (
	"api/internal/platform/response"
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Service struct {
	users UserProvider
}

func NewService(users UserProvider) *Service {
	return &Service{
		users: users,
	}
}

type UserProvider interface {
	PurgeUser(ctx context.Context, id uuid.UUID) (uuid.UUID, error)
	RestoreUser(ctx context.Context, id uuid.UUID) (uuid.UUID, string, error)
}

/* Hard delete user (with all albums via db onDelete) */
func (s *Service) PurgeUser(ctx context.Context, targetID uuid.UUID, adminID uuid.UUID) (uuid.UUID, error) {
	// Check if admin is not deleting himself
	if adminID == targetID {
		return uuid.Nil, response.ErrNoPermission
	}

	uID, err := s.users.PurgeUser(ctx, targetID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, response.ErrUserNotFound.Wrap(err)
		}
		return uuid.Nil, err
	}
	return uID, nil
}

/* Restore deleted user */
func (s *Service) RestoreUser(ctx context.Context, targetID uuid.UUID, adminID uuid.UUID) (uuid.UUID, string, error) {
	// Check if admin is not restoring himself
	if adminID == targetID {
		return uuid.Nil, "", response.ErrNoPermission
	}

	id, slug, err := s.users.RestoreUser(ctx, targetID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, "", response.ErrUserNotFound.Wrap(err)
		}
		return uuid.Nil, "", err
	}
	return id, slug, nil
}
