DROP TABLE IF EXISTS users;

CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    nickname VARCHAR(50) NOT NULL UNIQUE,
    age INTEGER NOT NULL,
    gender VARCHAR(10) NOT NULL,
    first_name VARCHAR(50) NOT NULL,
    last_name VARCHAR(50) NOT NULL,
    email VARCHAR(100) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

DROP TABLE IF EXISTS sessions;

CREATE TABLE IF NOT EXISTS sessions (
    user_id TEXT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    session_id TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP
);

DROP TABLE IF EXISTS posts;

CREATE TABLE IF NOT EXISTS posts (
    id           TEXT PRIMARY KEY,
    author_id    TEXT    NOT NULL,
    title        TEXT    NOT NULL CHECK(LENGTH(title) <= 255),
    content      TEXT    NOT NULL,
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(author_id)   REFERENCES users(id) ON DELETE CASCADE
);

DROP TABLE IF EXISTS comments;

CREATE TABLE IF NOT EXISTS comments (
    id                TEXT PRIMARY KEY,             
    post_id           TEXT NOT NULL,               
    author_id         TEXT NOT NULL,                
    content           TEXT NOT NULL,                 
    created_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(post_id) REFERENCES posts(id)   ON DELETE CASCADE,
    FOREIGN KEY(author_id) REFERENCES users(id) ON DELETE CASCADE
);

DROP TABLE IF EXISTS messages;

CREATE TABLE IF NOT EXISTS messages (
id INTEGER PRIMARY KEY AUTOINCREMENT,
from_user INTEGER NOT NULL,
to_user INTEGER NOT NULL,
body TEXT NOT NULL,
created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
FOREIGN KEY(from_user) REFERENCES users(id) ON DELETE CASCADE,
FOREIGN KEY(to_user) REFERENCES users(id) ON DELETE CASCADE
);

DROP TABLE IF EXISTS categories;

CREATE TABLE IF NOT EXISTS categories(
    id INTEGER PRIMARY KEY,
    name TEXT UNIQUE NOT NULL
);

DROP TABLE post_categories;

CREATE TABLE IF NOT EXISTS post_categories (
    post_id    TEXT    NOT NULL
               REFERENCES posts(id)       ON DELETE CASCADE,
    category_id INTEGER NOT NULL
               REFERENCES categories(id)  ON DELETE CASCADE,
    PRIMARY KEY (post_id, category_id)
);