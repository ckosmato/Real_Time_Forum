CREATE TABLE IF NOT EXISTS categories(
    id INTEGER PRIMARY KEY,
    name TEXT UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS post_categories (
    post_id    TEXT    NOT NULL
               REFERENCES posts(id)       ON DELETE CASCADE,
    category_id INTEGER NOT NULL
               REFERENCES categories(id)  ON DELETE CASCADE,
    PRIMARY KEY (post_id, category_id)
);