-- +goose Up
CREATE TABLE IF NOT EXISTS links (
    id SERIAL PRIMARY KEY,
    original_url TEXT NOT NULL,
    short_name VARCHAR(64) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_links_short_name ON links(short_name);

-- +goose Down
DROP INDEX IF EXISTS idx_links_short_name;
DROP TABLE IF EXISTS links;
