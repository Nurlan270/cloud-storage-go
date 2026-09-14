-- +goose Up
CREATE UNIQUE INDEX IF NOT EXISTS idx_resources_userid_path_name_type ON resources (user_id, path, name, type);

-- +goose Down
DROP INDEX IF EXISTS idx_resources_userid_path_name_type;
