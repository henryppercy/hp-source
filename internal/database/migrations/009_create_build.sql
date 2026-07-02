CREATE TABLE build (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    built_at TEXT NOT NULL DEFAULT (datetime('now')),
    location_id INTEGER REFERENCES location(id),
    go_version TEXT NOT NULL,
    built_on TEXT NOT NULL,
    success INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
