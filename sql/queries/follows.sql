-- name: CreateFollow :one
INSERT INTO
    follows (id, feed_id, user_id, created_at, updated_at)
VALUES
    ($1, $2, $3, $4, $5)
RETURNING
    *;

-- name: Unfollow :one
DELETE FROM follows
WHERE
    id = $1
    AND user_id = $2
RETURNING
    id;

-- name: GetFollows :many
SELECT
    *
FROM
    follows
where user_id = $1;