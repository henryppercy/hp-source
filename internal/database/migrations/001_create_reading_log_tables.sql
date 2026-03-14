-- Migration 001: Initial schema

CREATE TABLE author (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    sort_name TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE series (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE genre (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE book (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    headline TEXT,
    date_published TEXT NOT NULL,
    original_language TEXT NOT NULL DEFAULT 'english',
    type TEXT NOT NULL CHECK(type IN ('fiction', 'non-fiction')),
    genre_id INTEGER NOT NULL REFERENCES genre(id),
    series_id INTEGER REFERENCES series(id),
    series_position REAL,
    shelf_status TEXT,
    url TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE book_author (
    book_id INTEGER NOT NULL REFERENCES book(id) ON DELETE CASCADE,
    author_id INTEGER NOT NULL REFERENCES author(id),
    role TEXT NOT NULL DEFAULT 'author',
    PRIMARY KEY (book_id, author_id, role)
);

CREATE TABLE tag (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE book_tag (
    book_id INTEGER NOT NULL REFERENCES book(id) ON DELETE CASCADE,
    tag_id INTEGER NOT NULL REFERENCES tag(id) ON DELETE CASCADE,
    PRIMARY KEY (book_id, tag_id)
);

CREATE TABLE book_copy (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    book_id INTEGER NOT NULL REFERENCES book(id) ON DELETE CASCADE,
    format TEXT NOT NULL,
    page_count INTEGER,
    language TEXT NOT NULL DEFAULT 'english',
    isbn TEXT,
    cover_image TEXT,
    source TEXT,
    date_acquired TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE read (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    book_id INTEGER NOT NULL REFERENCES book(id) ON DELETE CASCADE,
    copy_id INTEGER REFERENCES book_copy(id),
    status TEXT NOT NULL DEFAULT 'reading' CHECK(status IN ('reading', 'finished', 'abandoned')),
    rating INTEGER CHECK(rating BETWEEN 1 AND 10),
    date_started TEXT,
    date_finished TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Triggers for updated_at

CREATE TRIGGER book_updated_at AFTER UPDATE ON book
BEGIN
    UPDATE book SET updated_at = datetime('now') WHERE id = NEW.id;
END;

CREATE TRIGGER read_updated_at AFTER UPDATE ON read
BEGIN
    UPDATE read SET updated_at = datetime('now') WHERE id = NEW.id;
END;

-- Genre seeds

INSERT INTO genre (name) VALUES
    ('literary'),
    ('thriller'),
    ('mystery'),
    ('science fiction'),
    ('fantasy'),
    ('horror'),
    ('romance'),
    ('historical'),
    ('adventure'),
    ('comedy'),
    ('short stories'),
    ('biography'),
    ('history'),
    ('science'),
    ('philosophy'),
    ('politics'),
    ('self-help');
