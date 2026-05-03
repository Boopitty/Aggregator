-- name: SearchUserID :one
SELECT * FROM users
WHERE id = $1;