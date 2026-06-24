package users

import (
	"api/internal/domain/shared/types"
	db "api/internal/platform/database/sqlc"
	"time"

	"github.com/google/uuid"
)

// Full user (stored in cache, standart type)

type User struct {
	ID        uuid.UUID  `json:"id"`
	Email     string     `json:"email"`
	Username  string     `json:"username"`
	Avatar    string     `json:"avatar"`
	Slug      string     `json:"slug"`
	Role      types.Role `json:"role"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

func FromDB(u db.User) User {
	var deletedAt *time.Time
	if u.DeletedAt.Valid {
		deletedAt = &u.DeletedAt.Time
	}
	return User{
		ID:        u.ID,
		Email:     u.Email,
		Username:  u.Username,
		Avatar:    u.Avatar,
		Slug:      u.Slug,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		DeletedAt: deletedAt,
	}
}

// Me User

type MeUser struct {
	ID        uuid.UUID  `json:"id"`
	Email     string     `json:"email"`
	Username  string     `json:"username"`
	Avatar    string     `json:"avatar"`
	Slug      string     `json:"slug"`
	Role      types.Role `json:"role"`
	CreatedAt time.Time  `json:"created_at"`
}

func ToMe(u User) MeUser {
	return MeUser{
		ID:        u.ID,
		Email:     u.Email,
		Username:  u.Username,
		Avatar:    u.Avatar,
		Slug:      u.Slug,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
	}
}

// Public user

type PublicUser struct {
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
	Slug     string `json:"slug"`
}

func ToPublic(u User) PublicUser {
	return PublicUser{
		Username: u.Username,
		Avatar:   u.Avatar,
		Slug:     u.Slug,
	}
}
