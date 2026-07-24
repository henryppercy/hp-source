ALTER TABLE book_copy ADD COLUMN title TEXT NOT NULL DEFAULT '';
ALTER TABLE book_copy ADD COLUMN headline TEXT;
ALTER TABLE book_copy ADD COLUMN shelf_status TEXT;

UPDATE book_copy SET
    title = (SELECT b.title FROM book b WHERE b.id = book_copy.book_id),
    headline = (SELECT b.headline FROM book b WHERE b.id = book_copy.book_id),
    shelf_status = (SELECT b.shelf_status FROM book b WHERE b.id = book_copy.book_id);

ALTER TABLE book DROP COLUMN headline;
ALTER TABLE book DROP COLUMN shelf_status;
