ALTER TABLE users ADD COLUMN book_collection_order TEXT[] NOT NULL DEFAULT '{}';
ALTER TABLE users ADD COLUMN course_collection_order TEXT[] NOT NULL DEFAULT '{}';
