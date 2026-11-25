ALTER TABLE book ADD COLUMN original_title TEXT;

UPDATE book SET original_title = title;
