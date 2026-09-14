-- +goose Up
ALTER TABLE users
    ALTER COLUMN id TYPE BIGINT;

ALTER SEQUENCE users_id_seq
    AS BIGINT;

-- +goose Down
ALTER TABLE users
    ALTER COLUMN id TYPE INTEGER;

ALTER SEQUENCE users_id_seq
    AS INTEGER;
