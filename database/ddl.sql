DROP TABLE IF EXISTS users;

CREATE TABLE IF NOT EXISTS users (
    id  VARCHAR(255) PRIMARY KEY,
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
    session_id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP
);

DROP TABLE IF EXISTS posts;

CREATE TABLE IF NOT EXISTS posts (
    id           VARCHAR(255) PRIMARY KEY,
    author_id    VARCHAR(255)    NOT NULL,
    title        VARCHAR(255)    NOT NULL,
    content      VARCHAR    NOT NULL,
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(author_id)   REFERENCES users(id) ON DELETE CASCADE
);

DROP TABLE IF EXISTS comments;

CREATE TABLE IF NOT EXISTS comments (
    id                VARCHAR(255) PRIMARY KEY,             
    post_id           VARCHAR(255) NOT NULL,               
    author_id         VARCHAR(255) NOT NULL,                
    content           VARCHAR(255) NOT NULL,                 
    created_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(post_id) REFERENCES posts(id)   ON DELETE CASCADE,
    FOREIGN KEY(author_id) REFERENCES users(id) ON DELETE CASCADE
);

DROP TABLE IF EXISTS messages;

CREATE TABLE IF NOT EXISTS messages (
id INTEGER PRIMARY KEY AUTOINCREMENT,
from_user INTEGER NOT NULL,
to_user INTEGER NOT NULL,
body VARCHAR NOT NULL,
created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
FOREIGN KEY(from_user) REFERENCES users(id) ON DELETE CASCADE,
FOREIGN KEY(to_user) REFERENCES users(id) ON DELETE CASCADE
);

DROP TABLE IF EXISTS categories;

CREATE TABLE IF NOT EXISTS categories(
    id INTEGER PRIMARY KEY,
    name VARCHAR(25) UNIQUE NOT NULL
);

DROP TABLE post_categories;

CREATE TABLE IF NOT EXISTS post_categories (
    post_id    VARCHAR(255)    NOT NULL
               REFERENCES posts(id)       ON DELETE CASCADE,
    category_id INTEGER NOT NULL
               REFERENCES categories(id)  ON DELETE CASCADE,
    PRIMARY KEY (post_id, category_id)
);