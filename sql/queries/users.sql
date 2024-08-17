-- name: CreateUser :one
INSERT INTO
    users (id, created_at, updated_at, NAME, api_key)
VALUES
    (
        $1,
        $2,
        $3,
        $4,
        ENCODE(SHA256(RANDOM() :: TEXT :: bytea), 'hex')
    ) RETURNING *;

-- name: GetUser :one
SELECT
    *
FROM
    users
WHERE
    api_key = $1;