-- users table
CREATE TABLE IF NOT EXISTS users (
    id           TEXT PRIMARY KEY,
    firebase_uid TEXT UNIQUE NOT NULL,
    email        TEXT UNIQUE NOT NULL,
    display_name TEXT NOT NULL,
    avatar_url   TEXT DEFAULT '',
    created_at   DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- groups table
CREATE TABLE IF NOT EXISTS groups (
    id                 INTEGER PRIMARY KEY AUTOINCREMENT,
    name               TEXT NOT NULL,
    description        TEXT DEFAULT '',
    created_by_user_id TEXT NOT NULL,
    created_at         DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (created_by_user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- group members  table
CREATE TABLE IF NOT EXISTS group_members (
    id        INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id  INTEGER NOT NULL,
    user_id   TEXT NOT NULL,
    role      TEXT NOT NULL DEFAULT 'member', -- 'admin' or 'member'
    joined_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id)  REFERENCES users(id)  ON DELETE CASCADE,
    UNIQUE(group_id, user_id)
);

-- expenses table
CREATE TABLE IF NOT EXISTS expenses (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id         INTEGER NOT NULL,
    paid_by_user_id  TEXT NOT NULL,
    title            TEXT NOT NULL,
    amount           REAL NOT NULL CHECK(amount > 0),
    currency         TEXT NOT NULL DEFAULT 'INR',
    category         TEXT DEFAULT 'general',
    split_type       TEXT NOT NULL,
    expense_date     DATETIME DEFAULT CURRENT_TIMESTAMP,
    created_at       DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (group_id)        REFERENCES groups(id) ON DELETE CASCADE,
    FOREIGN KEY (paid_by_user_id) REFERENCES users(id)  ON DELETE CASCADE
);

-- expense split expense among members
CREATE TABLE IF NOT EXISTS expense_splits (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    expense_id   INTEGER NOT NULL,
    user_id      TEXT NOT NULL,
    owed_amount  REAL NOT NULL CHECK(owed_amount >= 0),
    is_settled   INTEGER NOT NULL DEFAULT 0,
    settled_at   DATETIME,
    FOREIGN KEY (expense_id) REFERENCES expenses(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id)    REFERENCES users(id)    ON DELETE CASCADE
);
