CREATE TABLE IF NOT EXISTS posts (
    id           TEXT PRIMARY KEY,
    author_id    TEXT    NOT NULL,
    title        TEXT    NOT NULL CHECK(LENGTH(title) <= 255),
    content      TEXT    NOT NULL,
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_deleted   INTEGER NOT NULL DEFAULT 0,
    likes_count  INTEGER NOT NULL DEFAULT 0,
    dislikes_count INTEGER NOT NULL DEFAULT 0,
    image_name  TEXT  NOT NULL,
    status TEXT NOT NULL                                       
        CHECK(status IN ('pending','approved')) 
        DEFAULT 'pending',
    flag TEXT DEFAULT NULL,
    FOREIGN KEY(author_id)   REFERENCES users(id) ON DELETE CASCADE
);
