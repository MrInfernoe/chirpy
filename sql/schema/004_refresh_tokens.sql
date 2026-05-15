-- +goose Up
CREATE TABLE refresh_tokens (
    token       TEXT        PRIMARY KEY,
    created_at  TIMESTAMP   NOT NULL,
    updated_at  TIMESTAMP   NOT NULL,
    user_id     UUID        NOT NULL        REFERENCES users
                                                ON DELETE CASCADE
                                                ON UPDATE CASCADE,
    expires_at  TIMESTAMP   NOT NULL,
    revoked_at  TIMESTAMP,
    UNIQUE(token),
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- +goose Down
DROP TABLE refresh_tokens;