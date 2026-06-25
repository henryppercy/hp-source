CREATE TABLE post (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    slug TEXT NOT NULL UNIQUE,
    title TEXT NOT NULL,
    type TEXT NOT NULL DEFAULT '' CHECK(type IN ('', 'slice', 'spanish')),
    headline TEXT,
    body TEXT NOT NULL DEFAULT '',
    posted_at TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TRIGGER post_updated_at AFTER UPDATE ON post
BEGIN
    UPDATE post SET updated_at = datetime('now') WHERE id = NEW.id;
END;
