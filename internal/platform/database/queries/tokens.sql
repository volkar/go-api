-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (user_id, token_hash, expires_at, ip, ua, device, os, browser, location)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id;

-- name: GetRefreshTokenByHash :one
SELECT * FROM refresh_tokens
WHERE token_hash = $1
LIMIT 1;

-- name: ConsumeRefreshTokenByHash :execresult
UPDATE refresh_tokens
SET is_consumed = true, updated_at = NOW()
WHERE token_hash = $1 AND is_consumed = false;

-- name: ConsumeOtherRefreshTokensForUser :exec
UPDATE refresh_tokens
SET is_consumed = true, updated_at = NOW()
WHERE user_id = $1 AND token_hash != $2 AND is_consumed = false;

-- name: DeleteAllRefreshTokensForUser :exec
DELETE FROM refresh_tokens WHERE user_id = $1;

-- name: CleanupRefreshTokens :exec
DELETE FROM refresh_tokens
WHERE expires_at < NOW()
   OR (is_consumed = true AND updated_at < NOW() - INTERVAL '24 hours');