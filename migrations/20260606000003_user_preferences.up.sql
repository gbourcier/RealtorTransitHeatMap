CREATE TABLE user_preferences (
    user_id           UUID        PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    default_filter_id UUID        REFERENCES saved_filters(id) ON DELETE SET NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);
