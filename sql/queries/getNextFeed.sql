-- name: GetNextFeedToFetch :one
SELECT * FROM feeds
ORDER BY last_fetched_at Asc NULLS FIRST
LIMIT 1;