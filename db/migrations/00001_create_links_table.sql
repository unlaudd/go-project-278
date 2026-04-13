-- +goose Up
-- Creates the main table for storing shortened URL mappings.
CREATE TABLE IF NOT EXISTS links (
    id SERIAL PRIMARY KEY,
    original_url TEXT NOT NULL,           -- The target URL; TEXT allows arbitrary length
    short_name VARCHAR(64) UNIQUE NOT NULL, -- Short identifier; indexed for fast lookups
    created_at TIMESTAMPTZ DEFAULT NOW()    -- Timestamp with timezone for consistent serialization
);

-- Index on short_name to optimize redirect lookups (GET /r/:shortName).
CREATE INDEX IF NOT EXISTS idx_links_short_name ON links(short_name);

-- +goose Down
-- Reverts the migration: removes index and table.
DROP INDEX IF EXISTS idx_links_short_name;
DROP TABLE IF EXISTS links;
