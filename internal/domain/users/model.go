package users

import (
	"api/internal/domain/albums"
	db "api/internal/platform/database/sqlc"

	"github.com/google/uuid"
)

// Full user (stored in cache, standart type)

type User struct {
	ID       uuid.UUID `json:"id"`
	Email    string    `json:"email"`
	Username string    `json:"username"`
	Slug     string    `json:"slug"`
	Role     string    `json:"role"`
}

func FromDB(u db.User) User {
	return User{
		ID:       u.ID,
		Email:    u.Email,
		Username: u.Username,
		Slug:     u.Slug,
		Role:     u.Role,
	}
}

// Public user (returned to client)

type PublicUser struct {
	Username string `json:"username"`
	Slug     string `json:"slug"`
}

func ToPublic(u User) PublicUser {
	return PublicUser{
		Username: u.Username,
		Slug:     u.Slug,
	}
}

// Profile (user with albums)

type Profile struct {
	User   User                 `json:"user"`
	Albums []albums.AlbumInList `json:"albums"`
}

type PublicProfile struct {
	User   PublicUser                 `json:"user"`
	Albums []albums.PublicAlbumInList `json:"albums"`
}

func ToPublicProfile(p Profile) PublicProfile {
	return PublicProfile{
		User:   ToPublic(p.User),
		Albums: albums.ToPublicAlbumList(p.Albums),
	}
}
