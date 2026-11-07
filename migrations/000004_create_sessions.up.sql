CREATE TABLE sessions (
    id BIGSERIAL PRIMARY KEY,
    user_id	BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ
);

CREATE INDEX idx_sessions_user_id
ON sessions(user_id);