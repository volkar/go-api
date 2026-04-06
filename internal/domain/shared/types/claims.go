package types

import (
	"github.com/google/uuid"
)

// UserClaims represents the claims in the PASETO token
type UserClaims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Role   Role      `json:"role"`
}
