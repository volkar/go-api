package tokens

import (
	db "api/internal/platform/database/sqlc"
	"api/internal/platform/response"
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	q    db.Querier
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		q:    db.New(pool),
		pool: pool,
	}
}

/* Create token with given user id, hash and expiration time */
func (r *Repository) Create(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time, ip string, ua string, device string, os string, browser string, location string) (uuid.UUID, error) {
	return r.q.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
		Ip:        ip,
		Ua:        ua,
		Device:    device,
		Os:        os,
		Browser:   browser,
		Location:  location,
	})
}

/* Get token by hash */
func (r *Repository) GetByHash(ctx context.Context, hash string) (db.RefreshToken, error) {
	return r.q.GetRefreshTokenByHash(ctx, hash)
}

/* Consume one token by hash */
func (r *Repository) ConsumeByHash(ctx context.Context, hash string) error {
	_, err := r.q.ConsumeRefreshTokenByHash(ctx, hash)
	return err
}

/* Consume other tokens for given user id except given hash */
func (r *Repository) ConsumeOtherForUser(ctx context.Context, userID uuid.UUID, exceptHash string) error {
	return r.q.ConsumeOtherRefreshTokensForUser(ctx, db.ConsumeOtherRefreshTokensForUserParams{
		UserID:    userID,
		TokenHash: exceptHash,
	})
}

/* Delete all tokens for given user id */
func (r *Repository) DeleteAllRefreshForUser(ctx context.Context, userID uuid.UUID) error {
	return r.q.DeleteAllRefreshTokensForUser(ctx, userID)
}

/* Delete expired and consumed tokens */
func (r *Repository) Cleanup(ctx context.Context) error {
	return r.q.CleanupRefreshTokens(ctx)
}

/* Consume old token and create new one in single transaction */
func (r *Repository) ReplaceInTransaction(ctx context.Context, userID uuid.UUID, oldHash string, newHash string, expiresAt time.Time, ip string, ua string, device string, os string, browser string, location string) error {
	// Transaction pool and query
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	qtx := db.New(tx)

	// Consume old token
	result, err := qtx.ConsumeRefreshTokenByHash(ctx, oldHash)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return response.ErrTokenConsumed
	}

	// Create new token
	_, err = qtx.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
		UserID:    userID,
		TokenHash: newHash,
		ExpiresAt: expiresAt,
		Ip:        ip,
		Ua:        ua,
		Device:    device,
		Os:        os,
		Browser:   browser,
		Location:  location,
	})
	if err != nil {
		return err
	}

	// Commit transaction
	return tx.Commit(ctx)
}
