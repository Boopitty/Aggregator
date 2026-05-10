-- name: GetPostsForUser :many
SELECT * FROM posts
WHERE feed_id = $1
ORDER BY published_at DESC NULLS LAST
LIMIT $2;