-- +goose Up
CREATE TABLE refresh_tokens (
    token TEXT PRIMARY KEY NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    revoked_at TIMESTAMP,
    user_id UUID REFERENCES users ON DELETE CASCADE NOT NULL,
    CONSTRAINT fk_users_tokens
      FOREIGN KEY(user_id)
        REFERENCES users(id)
);

-- +goose Down
DROP TABLE refresh_tokens;
