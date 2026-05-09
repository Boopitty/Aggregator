-- name: UpdatePost :exec
UPDATE posts
SET updated_at = $2, title = $3, description = $4
WHERE feed_id = $1;