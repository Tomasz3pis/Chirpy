-- +goose Up
CREATE TABLE users (
    id UUID PRIMARY KEY NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    email VARCHAR(255) NOT NULL,
    hashed_password TEXT NOT NULL
);

-- +goose Down
DROP TABLE users;
