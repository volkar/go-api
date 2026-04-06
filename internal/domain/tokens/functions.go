package tokens

import (
	"api/internal/domain/shared/types"
	"context"
)

/* Insert user claims to context */
func InsertClaimsToContext(ctx context.Context, claims types.UserClaims) context.Context {
	return context.WithValue(ctx, userClaimsKey, claims)
}

/* Get user claims from context */
func GetClaimsFromContext(ctx context.Context) (types.UserClaims, bool) {
	claims, ok := ctx.Value(userClaimsKey).(types.UserClaims)
	return claims, ok
}
