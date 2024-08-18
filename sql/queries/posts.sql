-- name: CreatePost :one
INSERT INTO
    posts (
        id,
        created_at,
        updated_at,
        title,
        url,
        description,
        published_at,
        feed_id
    )
VALUES
    (
        $1,
        (NOW() AT TIME ZONE 'utc'),
        (NOW() AT TIME ZONE 'utc'),
        $2,
        $3,
        $4,
        $5,
        $6
    )
RETURNING
    *;

-- name: GetPostsByUser :many
SELECT
    posts.*
FROM
    posts
    INNER JOIN follows ON posts.feed_id = follows.feed_id
WHERE
    follows.user_id = $1
ORDER BY
    published_at
LIMIT
    $2;