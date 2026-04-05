-- name: GetAvailableAlbumBySlugs :one
SELECT sqlc.embed(a), sqlc.embed(u) FROM albums a
JOIN users u ON a.user_id = u.id
WHERE u.slug = @user_slug AND a.slug = @album_slug AND a.is_active AND a.deleted_at IS NULL AND u.deleted_at IS NULL AND (
      access = 'public'
      OR (access = 'shared' AND @viewer_email::text = ANY(shared_emails))
      OR @viewer_id::uuid = a.user_id::uuid
  );

-- name: ListAvailableAlbumIDs :many
SELECT id, date_at
FROM albums
WHERE user_id = @user_id
  AND deleted_at IS NULL
  AND (
      access = 'public'
      OR (access = 'shared' AND @viewer_email::text = ANY(shared_emails))
      OR @viewer_id::uuid = @user_id::uuid
  )
  AND (
      sqlc.narg('cursor_date_at')::timestamptz IS NULL
      OR date_at < sqlc.narg('cursor_date_at')::timestamptz
      OR (date_at = sqlc.narg('cursor_date_at')::timestamptz AND id < sqlc.narg('cursor_id')::uuid)
  )
ORDER BY date_at DESC, id DESC
LIMIT sqlc.arg('limit');

-- name: ListDeletedAlbumIDs :many
SELECT id, date_at
FROM albums
WHERE user_id = @user_id
  AND deleted_at IS NOT NULL
  AND (
      sqlc.narg('cursor_date_at')::timestamptz IS NULL
      OR date_at < sqlc.narg('cursor_date_at')::timestamptz
      OR (date_at = sqlc.narg('cursor_date_at')::timestamptz AND id < sqlc.narg('cursor_id')::uuid)
  )
ORDER BY date_at DESC, id DESC
LIMIT sqlc.arg('limit');

-- name: GetAlbumsByIDs :many
SELECT *
FROM albums a
WHERE id = ANY(@ids::uuid[]);

-- name: CreateAlbum :one
INSERT INTO albums (user_id, title, slug, atlas, access, shared_emails, date_at)
SELECT u.id, @title, @slug, @atlas, @access, @shared_emails, @date_at
FROM users u
WHERE u.id = @user_id AND u.deleted_at IS NULL
RETURNING *;

-- name: UpdateAlbum :one
WITH old_data AS (
  SELECT a.id, a.slug AS old_slug, u.slug AS user_slug
  FROM albums a
  JOIN users u ON a.user_id = u.id
  WHERE a.id = @album_id AND a.user_id = @user_id AND a.deleted_at IS NULL AND u.deleted_at IS NULL
  FOR UPDATE OF a
)
UPDATE albums
SET
  title = @title,
  slug = @slug,
  atlas = @atlas,
  access = @access,
  shared_emails = @shared_emails,
  date_at = @date_at,
  is_active = @is_active,
  updated_at = NOW()
FROM old_data
WHERE albums.id = old_data.id
RETURNING sqlc.embed(albums), old_data.old_slug, old_data.user_slug;

-- name: SoftDeleteAlbum :one
WITH old_data AS (
  SELECT a.id, u.slug AS user_slug
  FROM albums a
  JOIN users u ON a.user_id = u.id
  WHERE a.id = @album_id AND a.user_id = @user_id AND a.deleted_at IS NULL AND u.deleted_at IS NULL
  FOR UPDATE OF a
)
UPDATE albums
SET deleted_at = NOW(), updated_at = NOW()
FROM old_data
WHERE albums.id = old_data.id
RETURNING sqlc.embed(albums), old_data.user_slug;

-- name: RestoreAlbum :one
UPDATE albums a
SET
  deleted_at = NULL,
  updated_at = NOW(),
  slug = CASE
    WHEN EXISTS (
      SELECT 1 FROM albums AS a_active
      WHERE a_active.user_id = a.user_id
        AND a_active.slug = a.slug
        AND a_active.deleted_at IS NULL
        AND a_active.id != a.id
    )
    THEN left(a.slug, 246) || '-' || substring(md5(random()::text) from 1 for 8)
    ELSE a.slug
  END
FROM users u
WHERE a.user_id = u.id
  AND a.id = @album_id
  AND a.user_id = @user_id
  AND a.deleted_at IS NOT NULL
  AND u.deleted_at IS NULL
RETURNING a.id;

-- name: HardDeleteAlbum :one
DELETE FROM albums a
USING users u
WHERE a.user_id = u.id
  AND a.id = @album_id
  AND a.user_id = @user_id
  AND a.deleted_at IS NOT NULL
  AND u.deleted_at IS NULL
RETURNING a.id;