CREATE TABLE translator (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    sort_name TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE copy_translator (
    copy_id INTEGER NOT NULL REFERENCES book_copy(id) ON DELETE CASCADE,
    translator_id INTEGER NOT NULL REFERENCES translator(id),
    PRIMARY KEY (copy_id, translator_id)
);
