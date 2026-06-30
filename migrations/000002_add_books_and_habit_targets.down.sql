DROP TRIGGER IF EXISTS books_set_updated_at ON books;
DROP TABLE IF EXISTS books;
ALTER TABLE habits DROP COLUMN IF EXISTS monthly_target;
