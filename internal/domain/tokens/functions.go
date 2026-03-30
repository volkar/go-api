package tokens

import "context"

/* Insert user claims to context */
func InsertClaimsToContext(ctx context.Context, claims UserClaims) context.Context {
	return context.WithValue(ctx, userClaimsKey, claims)
}

/* Get user claims from context */
func GetClaimsFromContext(ctx context.Context) (UserClaims, bool) {
	claims, ok := ctx.Value(userClaimsKey).(UserClaims)
	return claims, ok
}
