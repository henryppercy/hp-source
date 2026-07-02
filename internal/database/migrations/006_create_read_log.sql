CREATE TABLE read_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    read_id INTEGER NOT NULL REFERENCES read(id) ON DELETE CASCADE,
    page INTEGER,
    note TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
