package users

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
	GetUser(ctx context.Context, userID uuid.UUID) (User, error)
	GetUserBySlug(ctx context.Context, userSlug string) (User, error)
	SetUser(ctx context.Context, user User)
	OnUserUpdated(ctx context.Context, userID uuid.UUID, userSlug string, oldUserSlug string)
	OnUserDeleted(ctx context.Context, userID uuid.UUID)
}

/* Upsert user by email and username, auth process */
func (r *Repository) Upsert(ctx context.Context, email string, username string) (User, error) {
	u, err := r.q.UpsertUser(ctx, db.UpsertUserParams{
		Email:    email,
		Username: username,
	})

	if err != nil {
		return User{}, err
	}

	user := FromDB(u)
	return user, nil
}

/* Create user */
func (r *Repository) Create(ctx context.Context, email string, username string, slug string, role string) (User, error) {
	u, err := r.q.CreateUser(ctx, db.CreateUserParams{
		Username: username,
		Slug:     slug,
		Email:    email,
		Role:     role,
	})

	if err != nil {
		return User{}, err
	}

	return FromDB(u), nil
}

/* Get non deleted user by id from cache */
func (r *Repository) GetAvailable(ctx context.Context, userID uuid.UUID) (User, error) {
	u, err := r.cache.GetUser(ctx, userID)
	if err == nil {
		// User found in cache, return
		return u, nil
	}

	// Not found in cache, get from the database
	dbUser, err := r.q.GetAvailableUser(ctx, userID)
	if err != nil {
		return User{}, err
	}
	user := FromDB(dbUser)

	// Set user with mapper to cache
	r.cache.SetUser(ctx, user)

	return user, nil
}

/* Get non deleted user by slug with cache */
func (r *Repository) GetAvailableBySlug(ctx context.Context, userSlug string) (User, error) {
	var user User

	// Try to get user from cache
	user, err := r.cache.GetUserBySlug(ctx, userSlug)

	// User not found in cache, get from database
	if user.ID == uuid.Nil {
		u, err := r.q.GetAvailableUserBySlug(ctx, userSlug)
		if err != nil {
			return User{}, err
		}

		user = FromDB(u)

		// Cache user with mapper to cache
		r.cache.SetUser(ctx, user)
	}

	if user.ID != uuid.Nil {
		return user, nil
	}

	return User{}, err
}

/* Update user */
func (r *Repository) Update(ctx context.Context, id uuid.UUID, username string, slug string) (User, error) {
	u, err := r.q.UpdateUser(ctx, db.UpdateUserParams{
		ID:       id,
		Username: username,
		Slug:     slug,
	})

	if err != nil {
		return User{}, err
	}

	// Call invalidate cache event
	r.cache.OnUserUpdated(ctx, id, u.User.Slug, u.OldSlug)

	return FromDB(u.User), nil
}

/* Delete user */
func (r *Repository) Delete(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	id, err := r.q.SoftDeleteUser(ctx, id)
	if err != nil {
		return uuid.UUID{}, err
	}

	// Call invalidate cache event
	r.cache.OnUserDeleted(ctx, id)

	return id, err
}
