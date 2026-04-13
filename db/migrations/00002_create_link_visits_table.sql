-- +goose Up
-- Creates the table for tracking link visit analytics.
-- Each record represents a single redirect event with client metadata.
CREATE TABLE IF NOT EXISTS link_visits (
    id SERIAL PRIMARY KEY,
    -- Foreign key to the links table; visits are automatically deleted if the link is removed.
    link_id INTEGER NOT NULL REFERENCES links(id) ON DELETE CASCADE,
    -- Client IP address; VARCHAR(45) accommodates both IPv4 (15 chars) and IPv6 (up to 45 chars).
    ip VARCHAR(45) NOT NULL,
    -- Optional client metadata for analytics and debugging.
    user_agent TEXT,
    referer TEXT,
    -- HTTP status code of the redirect response (e.g., 301, 302, 404).
    status SMALLINT NOT NULL,
    -- Timestamp with timezone for consistent serialization across environments.
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Index for querying visits by link (e.g., "show all visits for link X").
CREATE INDEX IF NOT EXISTS idx_link_visits_link_id ON link_visits(link_id);
-- Index for time-based queries: recent visits first, pagination, analytics dashboards.
CREATE INDEX IF NOT EXISTS idx_link_visits_created_at ON link_visits(created_at DESC);

-- +goose Down
-- Reverts the migration: removes indexes and table in dependency order.
DROP INDEX IF EXISTS idx_link_visits_created_at;
DROP INDEX IF EXISTS idx_link_visits_link_id;
DROP TABLE IF EXISTS link_visits;