ALTER TABLE users ADD COLUMN last_seen_at TIMESTAMPTZ;

UPDATE users u
SET last_seen_at = COALESCE(
    (SELECT max(created_at) FROM sessions s WHERE s.user_id = u.id),
    u.created_at
);

CREATE INDEX users_last_seen_at_idx ON users (last_seen_at);
