package users

import (
	db "api/internal/platform/database/sqlc"
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/singleflight"
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
	GetUser(ctx context.Context, userID uuid.UUID) (db.User, error)
	GetUserBySlug(ctx context.Context, userSlug string) (db.User, error)
	SetUser(ctx context.Context, user db.User)
	OnUserUpdated(ctx context.Context, user db.User, oldUserSlug string)
	OnUserDeleted(ctx context.Context, userID uuid.UUID)
}

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

/* Upsert user by email and username, auth process */
func (r *Repository) Upsert(ctx context.Context, email string, username string) (User, error) {
	u, err := r.q.UpsertUser(ctx, db.UpsertUserParams{
		Email:    email,
		Username: username,
	})
	if err != nil {
		return User{}, err
	}

	// Async set new user to cache
	bgCtx := context.WithoutCancel(ctx)
	go func(user db.User) {
		timeoutCtx, cancel := context.WithTimeout(bgCtx, 100*time.Millisecond)
		defer cancel()
		r.cache.SetUser(timeoutCtx, user)
	}(u)

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

	// Async set new user to cache
	bgCtx := context.WithoutCancel(ctx)
	go func(user db.User) {
		timeoutCtx, cancel := context.WithTimeout(bgCtx, 100*time.Millisecond)
		defer cancel()
		r.cache.SetUser(timeoutCtx, user)
	}(u)

	return FromDB(u), nil
}

/* Get non deleted user by id from cache */
func (r *Repository) GetAvailable(ctx context.Context, userID uuid.UUID) (User, error) {
	u, err := r.cache.GetUser(ctx, userID)
	if err == nil {
		// User found in cache, return
		return FromDB(u), nil
	}

	// Not found in cache, get from the database
	dbUser, err := r.q.GetAvailableUser(ctx, userID)
	if err != nil {
		return User{}, err
	}

	// Async set user to cache
	bgCtx := context.WithoutCancel(ctx)
	go func(user db.User) {
		timeoutCtx, cancel := context.WithTimeout(bgCtx, 100*time.Millisecond)
		defer cancel()
		r.cache.SetUser(timeoutCtx, user)
	}(dbUser)

	return FromDB(dbUser), nil
}

var userSlugGroup singleflight.Group

/* Get non deleted user by slug with cache */
func (r *Repository) GetAvailableBySlug(ctx context.Context, userSlug string) (User, error) {
	// Try to get user from cache
	user, _ := r.cache.GetUserBySlug(ctx, userSlug)
	if user.ID != uuid.Nil {
		return FromDB(user), nil
	}

	// User not found in cache, get from database
	// Use singleflight to prevent Cache Stampede
	sfKey := "sf:user:slug:" + userSlug
	val, sfErr, _ := userSlugGroup.Do(sfKey, func() (any, error) {
		dbUser, dbErr := r.q.GetAvailableUserBySlug(ctx, userSlug)
		if dbErr != nil {
			return db.User{}, dbErr
		}
		// Async set user to cache
		bgCtx := context.WithoutCancel(ctx)
		go func(u db.User) {
			timeoutCtx, cancel := context.WithTimeout(bgCtx, 100*time.Millisecond)
			defer cancel()
			r.cache.SetUser(timeoutCtx, u)
		}(dbUser)

		return dbUser, nil
	})

	if sfErr != nil {
		return User{}, sfErr
	}

	return FromDB(val.(db.User)), nil
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

	// Async call cache event
	bgCtx := context.WithoutCancel(ctx)
	go func(user db.User, oldUserSlug string) {
		timeoutCtx, cancel := context.WithTimeout(bgCtx, 100*time.Millisecond)
		defer cancel()
		r.cache.OnUserUpdated(timeoutCtx, user, oldUserSlug)
	}(u.User, u.OldSlug)

	return FromDB(u.User), nil
}

/* Delete user */
func (r *Repository) Delete(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	id, err := r.q.SoftDeleteUser(ctx, id)
	if err != nil {
		return uuid.UUID{}, err
	}

	// Async call cache event
	bgCtx := context.WithoutCancel(ctx)
	go func(userID uuid.UUID) {
		timeoutCtx, cancel := context.WithTimeout(bgCtx, 100*time.Millisecond)
		defer cancel()
		r.cache.OnUserDeleted(timeoutCtx, userID)
	}(id)

	return id, err
}
