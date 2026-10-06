-- +goose Up

ALTER TABLE users
ALTER COLUMN hashed_password SET NOT NULL;

-- +goose Down
ALTER TABLE users
ALTER COLUMN hashed_password DROP NOT NULL;