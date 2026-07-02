ALTER TABLE book_copy ADD COLUMN second_hand INTEGER NOT NULL DEFAULT 0;

UPDATE book_copy SET source = 'bought', second_hand = 1;
