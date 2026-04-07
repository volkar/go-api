-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS albums (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    title TEXT NOT NULL,
    cover TEXT NOT NULL DEFAULT '',
    date_at TIMESTAMPTZ NOT NULL,
    atlas JSONB NOT NULL DEFAULT '[]',
    access TEXT NOT NULL DEFAULT 'private',
    shared_emails TEXT[] NOT NULL DEFAULT '{}',
    direct_token UUID UNIQUE,
    slug TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ DEFAULT NULL
);
CREATE INDEX idx_albums_user_id ON albums(user_id);
CREATE UNIQUE INDEX idx_albums_user_slug_active ON albums (user_id, slug) WHERE (deleted_at IS NULL);
CREATE INDEX idx_albums_pagination ON albums (user_id, date_at DESC, id DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_albums_shared_emails ON albums USING GIN (shared_emails);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_albums_pagination;
DROP INDEX IF EXISTS idx_albums_user_slug_active;
DROP INDEX IF EXISTS idx_albums_user_id;
DROP INDEX IF EXISTS idx_albums_shared_emails;
DROP TABLE IF EXISTS albums;
-- +goose StatementEnd
