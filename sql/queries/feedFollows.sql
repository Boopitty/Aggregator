-- name: CreateFeedFollow :one
WITH inserted_follow AS (
    INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
    VALUES (
        $1,
        $2,
        $3,
        $4,
        $5
    )
    RETURNING *
)
SELECT inserted_follow.*,
       users.name AS user_name,
       feeds.name AS feed_name
FROM inserted_follow
INNER JOIN users ON inserted_follow.user_id = users.id
INNER JOIN feeds ON inserted_follow.feed_id = feeds.id;