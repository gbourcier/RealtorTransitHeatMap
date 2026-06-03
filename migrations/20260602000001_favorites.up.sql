CREATE TABLE favorites (
    user_id    UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    board      INTEGER     NOT NULL,
    mls        INTEGER     NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, board, mls),
    FOREIGN KEY (board, mls) REFERENCES listings (board, mls) ON DELETE CASCADE
);

CREATE INDEX favorites_user_created_idx ON favorites (user_id, created_at DESC);
