-- users table
CREATE TABLE IF NOT EXISTS users (
    id           TEXT PRIMARY KEY,
    firebase_uid TEXT UNIQUE NOT NULL,
    email        TEXT UNIQUE NOT NULL,
    display_name TEXT NOT NULL,
    avatar_url   TEXT DEFAULT '',
    created_at   DATETIME DEFAULT CURRENT_TIMESTAMP
);
