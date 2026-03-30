package albums

import (
	"api/internal/domain/albums/albumtypes"
	db "api/internal/platform/database/sqlc"
	"time"

	"github.com/google/uuid"
)

// Full album (stored in cache, standart type)

type Album struct {
	ID       uuid.UUID         `json:"id"`
	UserID   uuid.UUID         `json:"user_id"`
	Title    string            `json:"title"`
	Slug     string            `json:"slug"`
	Atlas    albumtypes.Atlas  `json:"atlas"`
	Access   albumtypes.Access `json:"access"`
	DateAt   time.Time         `json:"date_at"`
	IsActive bool              `json:"is_active"`
}

func FromDB(a db.Album) Album {
	return Album{
		ID:       a.ID,
		UserID:   a.UserID,
		Title:    a.Title,
		Slug:     a.Slug,
		Atlas:    a.Atlas,
		Access:   a.Access,
		DateAt:   a.DateAt,
		IsActive: a.IsActive,
	}
}

// Raw album in list (stored in cache, standart type)

type AlbumInList struct {
	ID       uuid.UUID         `json:"id"`
	Title    string            `json:"title"`
	Slug     string            `json:"slug"`
	Access   albumtypes.Access `json:"access"`
	DateAt   time.Time         `json:"date_at"`
	IsActive bool              `json:"is_active"`
}

func AlbumListAvailableFromDB(albums []db.ListAvailableAlbumsRow) []AlbumInList {
	albumsResponse := make([]AlbumInList, len(albums))
	for i := range albums {
		albumsResponse[i] = AlbumInList{
			ID:       albums[i].ID,
			Title:    albums[i].Title,
			Slug:     albums[i].Slug,
			Access:   albums[i].Access,
			DateAt:   albums[i].DateAt,
			IsActive: albums[i].IsActive,
		}
	}
	return albumsResponse
}

func AlbumListDeletedFromDB(albums []db.ListDeletedAlbumsRow) []AlbumInList {
	albumsResponse := make([]AlbumInList, len(albums))
	for i := range albums {
		albumsResponse[i] = AlbumInList{
			ID:       albums[i].ID,
			Title:    albums[i].Title,
			Slug:     albums[i].Slug,
			Access:   albums[i].Access,
			DateAt:   albums[i].DateAt,
			IsActive: albums[i].IsActive,
		}
	}
	return albumsResponse
}

// Public album (returned to client)

type PublicAlbum struct {
	Title  string           `json:"title"`
	Slug   string           `json:"slug"`
	Atlas  albumtypes.Atlas `json:"atlas"`
	DateAt time.Time        `json:"date_at"`
}

func ToPublic(a Album) PublicAlbum {
	return PublicAlbum{
		Title:  a.Title,
		Slug:   a.Slug,
		Atlas:  a.Atlas,
		DateAt: a.DateAt,
	}
}

// Public album in list (returned to client)

type PublicAlbumInList struct {
	Title  string    `json:"title"`
	Slug   string    `json:"slug"`
	DateAt time.Time `json:"date_at"`
}

func ToPublicAlbumList(albums []AlbumInList) []PublicAlbumInList {
	albumsResponse := make([]PublicAlbumInList, len(albums))
	for i := range albums {
		albumsResponse[i] = PublicAlbumInList{
			Title:  albums[i].Title,
			Slug:   albums[i].Slug,
			DateAt: albums[i].DateAt,
		}
	}
	return albumsResponse
}
