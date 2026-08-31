-- +goose Up
CREATE TABLE IF NOT EXISTS users
(
    id       SERIAL PRIMARY KEY,
    username VARCHAR(30) NOT NULL UNIQUE,
    password VARCHAR     NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS users;
