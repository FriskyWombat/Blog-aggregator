-- name: CreateUser :one
INSERT INTO
    users (id, created_at, updated_at, NAME, api_key)
VALUES
    (
        $1,
        (NOW() AT TIME ZONE 'utc'),
        (NOW() AT TIME ZONE 'utc'),
        $2,
        ENCODE(SHA256(RANDOM() :: TEXT :: bytea), 'hex')
    ) RETURNING *;

-- name: GetUser :one
SELECT
    *
FROM
    users
WHERE
    api_key = $1;