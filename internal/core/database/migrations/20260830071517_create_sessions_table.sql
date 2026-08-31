-- +goose Up
CREATE TABLE IF NOT EXISTS sessions
(
    uuid       VARCHAR PRIMARY KEY,
    user_id    INTEGER   NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    expires_at TIMESTAMP NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS sessions;
