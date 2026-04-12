-- name: CreateLink :one
INSERT INTO links (original_url, short_name)
VALUES ($1, $2)
RETURNING id, original_url, short_name, created_at;

-- name: GetLinkByID :one
SELECT id, original_url, short_name, created_at
FROM links
WHERE id = $1;

-- name: GetLinkByShortName :one
SELECT id, original_url, short_name, created_at
FROM links
WHERE short_name = $1;

-- name: ListLinks :many
SELECT id, original_url, short_name, created_at
FROM links
ORDER BY id DESC
LIMIT $1 OFFSET $2;

-- name: UpdateLink :one
UPDATE links
SET original_url = COALESCE($2, original_url),
    short_name = COALESCE($3, short_name)
WHERE id = $1
RETURNING id, original_url, short_name, created_at;

-- name: DeleteLink :exec
DELETE FROM links WHERE id = $1;
