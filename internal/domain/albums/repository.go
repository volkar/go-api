package albums

import (
	"api/internal/domain/albums/albumtypes"
	db "api/internal/platform/database/sqlc"
	"context"
	"time"

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
	GetAlbum(ctx context.Context, userSlug string, albumSlug string) (Album, error)
	SetAlbum(ctx context.Context, userSlug string, albumSlug string, album Album)
	GetAvailableAlbums(ctx context.Context, userID uuid.UUID) ([]AlbumInList, error)
	SetAvailableAlbums(ctx context.Context, userID uuid.UUID, albums []AlbumInList)
	GetDeletedAlbums(ctx context.Context, userID uuid.UUID) ([]AlbumInList, error)
	SetDeletedAlbums(ctx context.Context, userID uuid.UUID, albums []AlbumInList)
	OnAlbumCreated(ctx context.Context, userID uuid.UUID)
	OnAlbumUpdated(ctx context.Context, userID uuid.UUID, userSlug string, oldAlbumSlug string)
	OnAlbumDeleted(ctx context.Context, userID uuid.UUID, userSlug string, oldAlbumSlug string)
	OnAllAlbumsDeleted(ctx context.Context, userID uuid.UUID)
	OnAlbumRestored(ctx context.Context, userID uuid.UUID)
	OnAlbumPurged(ctx context.Context, userID uuid.UUID)
}

/* Get available album by user slug and album slug from cache */
func (r *Repository) GetAvailable(ctx context.Context, userSlug string, albumSlug string) (Album, error) {
	// Get album from cache
	a, err := r.cache.GetAlbum(ctx, userSlug, albumSlug)
	if err == nil {
		// Album found in cache, return
		return a, nil
	}

	// Not found in cache, get album from database
	res, err := r.q.GetAvailableAlbum(ctx, db.GetAvailableAlbumParams{
		UserSlug:  userSlug,
		AlbumSlug: albumSlug,
	})
	if err != nil {
		return Album{}, err
	}
	album := FromDB(res.Album)

	// Set album to cache
	r.cache.SetAlbum(ctx, userSlug, albumSlug, album)

	return album, nil
}

/* Get list of all albums by user id from cache */
func (r *Repository) ListAvailable(ctx context.Context, userID uuid.UUID) ([]AlbumInList, error) {
	a, err := r.cache.GetAvailableAlbums(ctx, userID)
	if err == nil {
		// Found in cache, return
		return a, nil
	}

	// Not found in cache, get available albums from database
	dbAlbums, err := r.q.ListAvailableAlbums(ctx, userID)
	if err != nil {
		return []AlbumInList{}, err
	}
	albums := AlbumListAvailableFromDB(dbAlbums)

	// Set album list to cache
	r.cache.SetAvailableAlbums(ctx, userID, albums)

	return albums, nil
}

/* Get list of deleted albums by user id  */
func (r *Repository) ListDeleted(ctx context.Context, userID uuid.UUID) ([]AlbumInList, error) {
	a, err := r.cache.GetDeletedAlbums(ctx, userID)
	if err == nil {
		// Found in cache, return
		return a, nil
	}

	// Not found in cache, get from the DB
	dbAlbums, err := r.q.ListDeletedAlbums(ctx, userID)
	if err != nil {
		return []AlbumInList{}, err
	}
	albums := AlbumListDeletedFromDB(dbAlbums)

	// Set album to cache
	r.cache.SetDeletedAlbums(ctx, userID, albums)

	return albums, nil
}

/* Create album */
func (r *Repository) Create(ctx context.Context, title string, slug string, atlas albumtypes.Atlas, access albumtypes.Access, dateAt time.Time, userID uuid.UUID) (Album, error) {
	a, err := r.q.CreateAlbum(ctx, db.CreateAlbumParams{
		UserID: userID,
		Title:  title,
		Slug:   slug,
		Atlas:  atlas,
		Access: access,
		DateAt: dateAt,
	})

	if err != nil {
		return Album{}, err
	}

	// Call invalidate cache event
	r.cache.OnAlbumCreated(ctx, userID)

	// Map and return
	album := FromDB(a)
	return album, nil
}

/* Update album */
func (r *Repository) Update(ctx context.Context, userID uuid.UUID, albumID uuid.UUID, title string, slug string, atlas albumtypes.Atlas, access albumtypes.Access, dateAt time.Time, IsActive bool) (Album, error) {
	a, err := r.q.UpdateAlbum(ctx, db.UpdateAlbumParams{
		AlbumID:  albumID,
		UserID:   userID,
		Title:    title,
		Slug:     slug,
		Atlas:    atlas,
		Access:   access,
		DateAt:   dateAt,
		IsActive: IsActive,
	})

	if err != nil {
		return Album{}, err
	}

	// Call invalidate cache event
	r.cache.OnAlbumUpdated(ctx, userID, a.UserSlug, a.OldSlug)

	return FromDB(a.Album), err
}

/* Delete album */
func (r *Repository) Delete(ctx context.Context, userID uuid.UUID, albumID uuid.UUID) (uuid.UUID, error) {
	a, err := r.q.SoftDeleteAlbum(ctx, db.SoftDeleteAlbumParams{
		AlbumID: albumID,
		UserID:  userID,
	})

	if err != nil {
		return uuid.Nil, err
	}

	// Call invalidate cache event
	r.cache.OnAlbumDeleted(ctx, userID, a.UserSlug, a.Album.Slug)

	return a.Album.ID, nil
}

/* Delete all albums by user id */
func (r *Repository) DeleteAll(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	uuids, err := r.q.SoftDeleteAllAlbums(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Call invalidate cache event
	r.cache.OnAllAlbumsDeleted(ctx, userID)

	return uuids, nil
}

/* Restore deleted album */
func (r *Repository) Restore(ctx context.Context, userID uuid.UUID, albumID uuid.UUID) (uuid.UUID, error) {
	id, err := r.q.RestoreAlbum(ctx, db.RestoreAlbumParams{
		AlbumID: albumID,
		UserID:  userID,
	})
	if err != nil {
		return uuid.UUID{}, err
	}

	// Call invalidate cache event
	r.cache.OnAlbumRestored(ctx, userID)

	return id, nil
}

/* Purge deleted album */
func (r *Repository) Purge(ctx context.Context, userID uuid.UUID, albumID uuid.UUID) (uuid.UUID, error) {
	id, err := r.q.HardDeleteAlbum(ctx, db.HardDeleteAlbumParams{
		AlbumID: albumID,
		UserID:  userID,
	})

	if err != nil {
		return uuid.UUID{}, err
	}

	// Call invalidate cache event
	r.cache.OnAlbumPurged(ctx, userID)

	return id, nil
}
