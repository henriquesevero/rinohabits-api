UPDATE books SET status = 'quero_ler' WHERE status = 'na_estante';
ALTER TABLE books DROP CONSTRAINT IF EXISTS books_status_check;
ALTER TABLE books ADD CONSTRAINT books_status_check CHECK (status IN ('quero_ler', 'lendo', 'lido'));
