package admin

import (
	db "api/internal/platform/database/sqlc"
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	q     db.Querier
	cache Cacher
}

func NewRepository(pool *pgxpool.Pool, cache Cacher) *Repository {
	return &Repository{
		q:     db.New(pool),
		cache: cache,
	}
}

type Cacher interface {
	OnUserPurged(ctx context.Context, userID uuid.UUID)
}

/* Hard delete user (with all albums via db onDelete) */
func (r *Repository) PurgeUser(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	id, err := r.q.HardDeleteUser(ctx, id)
	if err != nil {
		return uuid.UUID{}, err
	}

	// Call invalidate cache event
	r.cache.OnUserPurged(ctx, id)

	return id, err
}

/* Restore deleted user */
func (r *Repository) RestoreUser(ctx context.Context, id uuid.UUID) (uuid.UUID, string, error) {
	res, err := r.q.RestoreUser(ctx, id)
	if err != nil {
		return uuid.Nil, "", err
	}
	return res.ID, res.Slug, nil
}
