CREATE TABLE spanish_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date TEXT NOT NULL,
    seconds INTEGER NOT NULL,
    activity TEXT NOT NULL DEFAULT 'ci' CHECK(activity IN ('ci', 'lesson', 'conversation')),
    source TEXT NOT NULL CHECK(source IN ('dreaming_spanish', 'manual')),
    note TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX spanish_log_date ON spanish_log(date);
