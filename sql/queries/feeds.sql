-- name: CreateFeed :one
INSERT INTO
    feeds (
        id,
        created_at,
        updated_at,
        NAME,
        url,
        user_id,
        last_fetched_at
    )
VALUES
    (
        $1,
        (NOW() AT TIME ZONE 'utc'),
        (NOW() AT TIME ZONE 'utc'),
        $2,
        $3,
        $4,
        (NOW() AT TIME ZONE 'utc')
    )
RETURNING
    *;

-- name: GetFeeds :many
SELECT
    *
FROM
    feeds;

-- name: GetNextFeedsToFetch :many
SELECT
    *
FROM
    feeds
ORDER BY
    last_fetched_at
LIMIT
    $1;

-- name: MarkFeedFetched :exec
UPDATE feeds
SET
    last_fetched_at = (NOW() AT TIME ZONE 'utc'),
    updated_at = (NOW() AT TIME ZONE 'utc')
WHERE
    id = $1;