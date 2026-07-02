CREATE TABLE location (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    slug TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    country_code TEXT,
    lat REAL,
    lng REAL,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE(name, country_code)
);

INSERT INTO location (slug, name, country_code, lat, lng) VALUES
    ('sheffield', 'Sheffield', 'GB', 53.22, -1.28),
    ('london', 'London', 'GB', 51.51, -0.13),
    ('gloucester', 'Gloucester', 'GB', 51.86, -2.24);

ALTER TABLE post ADD COLUMN location_id INTEGER REFERENCES location(id);
