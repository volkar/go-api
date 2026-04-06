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
func (s *Service) PurgeUser(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	uID, err := s.users.PurgeUser(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, response.ErrUserNotFound.Wrap(err)
		}
		return uuid.Nil, err
	}
	return uID, nil
}

/* Restore deleted user */
func (s *Service) RestoreUser(ctx context.Context, id uuid.UUID) (uuid.UUID, string, error) {
	id, slug, err := s.users.RestoreUser(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, "", response.ErrUserNotFound.Wrap(err)
		}
		return uuid.Nil, "", err
	}
	return id, slug, nil
}
