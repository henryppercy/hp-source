ALTER TABLE post ADD COLUMN kind TEXT NOT NULL DEFAULT 'article'
    CHECK(kind IN ('article', 'slice'));
UPDATE post SET kind = 'slice' WHERE type = 'slice';

CREATE TABLE topic (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE post_topic (
    post_id  INTEGER NOT NULL REFERENCES post(id) ON DELETE CASCADE,
    topic_id INTEGER NOT NULL REFERENCES topic(id) ON DELETE CASCADE,
    PRIMARY KEY (post_id, topic_id)
);

INSERT INTO topic (name) VALUES
    ('spanish'), ('personal'), ('tech'), ('productivity'),
    ('music'), ('travel'), ('food'), ('hiking'), ('reading');

INSERT INTO post_topic (post_id, topic_id)
SELECT p.id, t.id FROM post p JOIN topic t ON t.name = 'spanish'
WHERE p.type = 'spanish';

ALTER TABLE post DROP COLUMN type;
