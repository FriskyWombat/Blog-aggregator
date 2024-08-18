-- +goose Up
ALTER TABLE feeds
ADD last_fetched_at TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'utc');

-- +goose Down
ALTER TABLE feeds
DROP COLUMN last_fetched_at;