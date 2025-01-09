-- +goose Up
CREATE TABLE chirps (
    id UUID PRIMARY KEY NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    body VARCHAR(255) NOT NULL,
    user_id UUID REFERENCES users ON DELETE CASCADE NOT NULL,
    CONSTRAINT fk_users
      FOREIGN KEY(user_id)
        REFERENCES users(id)
);

-- +goose Down
DROP TABLE chirps;
