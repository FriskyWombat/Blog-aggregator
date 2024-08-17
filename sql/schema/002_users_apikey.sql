-- +goose Up
ALTER TABLE users
ADD api_key VARCHAR(64) UNIQUE NOT NULL DEFAULT ENCODE(SHA256(RANDOM()::TEXT::bytea), 'hex');

-- +goose Down
ALTER TABLE users
DROP COLUMN api_key;