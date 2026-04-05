package albums

import (
	"api/internal/domain/albums/albumtypes"
	"api/internal/platform/cursor"
	db "api/internal/platform/database/sqlc"
	"api/internal/platform/response"
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/singleflight"
)

type Repository struct {
	q             db.Querier
	cache         Cacher
	cursorManager *cursor.Cursor
}

func NewRepository(pool *pgxpool.Pool, cache Cacher, cursorManager *cursor.Cursor) *Repository {
	return &Repository{
		q:             db.New(pool),
		cache:         cache,
		cursorManager: cursorManager,
	}
}

type Cacher interface {
	GetAlbumBySlugs(ctx context.Context, userSlug string, albumSlug string) (db.Album, error)
	GetAlbumList(ctx context.Context, keys []uuid.UUID) (map[uuid.UUID]db.Album, error)
	SetAlbum(ctx context.Context, album db.Album)
	SetUser(ctx context.Context, user db.User)
	OnAlbumCreated(ctx context.Context, album db.Album)
	OnAlbumUpdated(ctx context.Context, album db.Album, oldSlug string)
	OnAlbumDeleted(ctx context.Context, album db.Album)
}

type albumPaginationRow struct {
	ID     uuid.UUID
	DateAt time.Time
}

var albumSlugsGroup singleflight.Group

/* Get available album by user slug and album slug from cache */
func (r *Repository) GetAvailable(ctx context.Context, userSlug string, albumSlug string, viewerID uuid.UUID, viewerEmail string) (Album, error) {
	// Get album from cache
	a, err := r.cache.GetAlbumBySlugs(ctx, userSlug, albumSlug)
	if err == nil {
		// Album found in cache, return
		return FromDB(a), nil
	}

	// Not found in cache, get album from database.
	// Use singleflight to prevent Cache Stampede
	sfKey := "sf:album:slugs:" + userSlug + ":" + albumSlug
	val, err, _ := albumSlugsGroup.Do(sfKey, func() (any, error) {
		res, dbErr := r.q.GetAvailableAlbumBySlugs(ctx, db.GetAvailableAlbumBySlugsParams{
			UserSlug:    userSlug,
			AlbumSlug:   albumSlug,
			ViewerEmail: viewerEmail,
			ViewerID:    viewerID,
		})
		if dbErr != nil {
			return Album{}, err
		}

		// Async set user and album to cache
		bgCtx := context.WithoutCancel(ctx)
		go func(user db.User, album db.Album) {
			timeoutCtx, cancel := context.WithTimeout(bgCtx, 100*time.Millisecond)
			defer cancel()
			r.cache.SetAlbum(timeoutCtx, album)
			r.cache.SetUser(timeoutCtx, user)
		}(res.User, res.Album)

		return FromDB(res.Album), nil
	})

	if err != nil {
		return Album{}, err
	}

	return val.(Album), nil
}

/* Get list of paginated albums by user id */
func (r *Repository) ListAvailable(ctx context.Context, userID uuid.UUID, viewerID uuid.UUID, viewerEmail string, cursor string, limit int32) ([]Album, string, error) {
	// Parse secure cursor
	cursorDate, cursorID, err := r.cursorManager.Parse(cursor)
	if err != nil {
		return []Album{}, "", response.ErrInvalidCursor.Wrap(err)
	}

	// Fetch list of IDs from database
	fetchLimit := limit + 1
	dbRows, err := r.q.ListAvailableAlbumIDs(ctx, db.ListAvailableAlbumIDsParams{
		UserID:       userID,
		ViewerID:     viewerID,
		ViewerEmail:  viewerEmail,
		CursorDateAt: cursorDate,
		CursorID:     cursorID,
		Limit:        fetchLimit,
	})
	if err != nil {
		return []Album{}, "", err
	}

	// Map sqlc specific type to our common pagination row
	idRows := make([]albumPaginationRow, len(dbRows))
	for i, row := range dbRows {
		idRows[i] = albumPaginationRow{
			ID:     row.ID,
			DateAt: row.DateAt,
		}
	}

	return r.hydrateAlbumsList(ctx, idRows, limit)
}

/* Get list of paginated albums by user id */
func (r *Repository) ListDeleted(ctx context.Context, userID uuid.UUID, cursor string, limit int32) ([]Album, string, error) {
	// Parse secure cursor
	cursorDate, cursorID, err := r.cursorManager.Parse(cursor)
	if err != nil {
		return []Album{}, "", response.ErrInvalidCursor.Wrap(err)
	}

	// Fetch list of IDs from database
	fetchLimit := limit + 1
	dbRows, err := r.q.ListDeletedAlbumIDs(ctx, db.ListDeletedAlbumIDsParams{
		UserID:       userID,
		CursorDateAt: cursorDate,
		CursorID:     cursorID,
		Limit:        fetchLimit,
	})
	if err != nil {
		return []Album{}, "", err
	}

	// Map sqlc specific type to our common pagination row
	idRows := make([]albumPaginationRow, len(dbRows))
	for i, row := range dbRows {
		idRows[i] = albumPaginationRow{
			ID:     row.ID,
			DateAt: row.DateAt,
		}
	}

	return r.hydrateAlbumsList(ctx, idRows, limit)
}

/* Create album */
func (r *Repository) Create(ctx context.Context, title string, slug string, atlas albumtypes.Atlas, access string, sharedEmails []string, dateAt time.Time, userID uuid.UUID) (Album, error) {
	a, err := r.q.CreateAlbum(ctx, db.CreateAlbumParams{
		UserID:       userID,
		Title:        title,
		Slug:         slug,
		Atlas:        atlas,
		Access:       access,
		SharedEmails: sharedEmails,
		DateAt:       dateAt,
	})
	if err != nil {
		return Album{}, err
	}

	// Async call cache event
	bgCtx := context.WithoutCancel(ctx)
	go func(album db.Album) {
		timeoutCtx, cancel := context.WithTimeout(bgCtx, 100*time.Millisecond)
		defer cancel()
		r.cache.OnAlbumCreated(timeoutCtx, a)
	}(a)

	// Map and return
	album := FromDB(a)
	return album, nil
}

/* Update album */
func (r *Repository) Update(ctx context.Context, userID uuid.UUID, albumID uuid.UUID, title string, slug string, atlas albumtypes.Atlas, access string, sharedEmails []string, dateAt time.Time, IsActive bool) (Album, error) {
	a, err := r.q.UpdateAlbum(ctx, db.UpdateAlbumParams{
		AlbumID:      albumID,
		UserID:       userID,
		Title:        title,
		Slug:         slug,
		Atlas:        atlas,
		Access:       access,
		SharedEmails: sharedEmails,
		DateAt:       dateAt,
		IsActive:     IsActive,
	})
	if err != nil {
		return Album{}, err
	}

	// Async call cache event
	bgCtx := context.WithoutCancel(ctx)
	go func(album db.Album, oldSlug string) {
		timeoutCtx, cancel := context.WithTimeout(bgCtx, 100*time.Millisecond)
		defer cancel()
		r.cache.OnAlbumUpdated(timeoutCtx, album, oldSlug)
	}(a.Album, a.OldSlug)

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

	// Async call cache event
	bgCtx := context.WithoutCancel(ctx)
	go func(album db.Album) {
		timeoutCtx, cancel := context.WithTimeout(bgCtx, 100*time.Millisecond)
		defer cancel()
		r.cache.OnAlbumDeleted(timeoutCtx, album)
	}(a.Album)

	return a.Album.ID, nil
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

	return id, nil
}

func (r *Repository) hydrateAlbumsList(ctx context.Context, idRows []albumPaginationRow, limit int32) ([]Album, string, error) {
	if len(idRows) == 0 {
		return []Album{}, "", nil
	}

	fetchLimit := limit + 1

	// Calculate next cursor if needed
	var nextCursor string
	if len(idRows) == int(fetchLimit) {
		idRows = idRows[:limit]
		lastItem := idRows[len(idRows)-1]
		nextCursor, _ = r.cursorManager.Encode(lastItem.DateAt, lastItem.ID.String())
	}

	// Prepare list of IDs
	albumIDs := make([]uuid.UUID, len(idRows))
	for i, row := range idRows {
		albumIDs[i] = row.ID
	}

	// Get albums from cache
	albums, _ := r.cache.GetAlbumList(ctx, albumIDs)

	resultMap := make(map[uuid.UUID]db.Album, len(idRows))
	var missingIDs []uuid.UUID

	// Separate cached and missing albums
	for _, uid := range albumIDs {
		if albums != nil && albums[uid].ID != uuid.Nil {
			resultMap[uid] = albums[uid]
		} else {
			missingIDs = append(missingIDs, uid)
		}
	}

	// Process missing albums
	if len(missingIDs) > 0 {
		dbAlbums, err := r.q.GetAlbumsByIDs(ctx, missingIDs)
		if err != nil {
			return []Album{}, "", err
		}

		// Detach context from parent
		bgCtx := context.WithoutCancel(ctx)

		for _, dbAlbum := range dbAlbums {
			resultMap[dbAlbum.ID] = dbAlbum
			// Async set missing albums to cache
			go func(a db.Album, bCtx context.Context) {
				timeoutCtx, cancel := context.WithTimeout(bCtx, 100*time.Millisecond)
				defer cancel()
				r.cache.SetAlbum(timeoutCtx, a)
			}(dbAlbum, bgCtx)
		}
	}

	// Restore strict ordering
	finalAlbums := make([]Album, 0, len(idRows))
	for _, row := range idRows {
		if album, ok := resultMap[row.ID]; ok {
			finalAlbums = append(finalAlbums, FromDB(album))
		}
	}

	return finalAlbums, nextCursor, nil
}
