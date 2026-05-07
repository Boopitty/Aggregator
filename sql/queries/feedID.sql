-- name: SearchFeedID :one
SELECT * FROM feeds WHERE id = $1;
