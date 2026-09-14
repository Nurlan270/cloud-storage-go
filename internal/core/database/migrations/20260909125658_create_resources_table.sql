-- +goose Up
CREATE TABLE IF NOT EXISTS resources
(
    id      BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users (id) ON DELETE CASCADE,
    path    VARCHAR NOT NULL,
    name    VARCHAR NOT NULL,
    size    BIGINT  NULL,
    type    VARCHAR NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS resources;
