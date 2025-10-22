CREATE TABLE IF NOT EXISTS comments (
    id                TEXT PRIMARY KEY,               -- uuid v4
    post_id           TEXT NOT NULL,                  -- FK → posts.id
    author_id         TEXT NOT NULL,                  -- FK → users.id
    content           TEXT NOT NULL,                  -- raw markdown / plain text
    likes_count       INTEGER NOT NULL DEFAULT 0,
    dislikes_count    INTEGER NOT NULL DEFAULT 0,
    is_deleted        INTEGER NOT NULL DEFAULT 0,     -- soft-delete flag
    created_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY(post_id) REFERENCES posts(id)   ON DELETE CASCADE,
    FOREIGN KEY(author_id) REFERENCES users(id) ON DELETE CASCADE
);