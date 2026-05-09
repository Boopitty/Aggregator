-- name: GetNextFeedToFetch :one
SELECT * FROM feeds
WHERE last_fetched_at IS NULL OR last_fetched_at < updated_at
ORDER BY last_fetched_at ASC NULLS FIRST
LIMIT 1;