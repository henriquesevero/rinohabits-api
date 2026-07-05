ALTER TABLE books DROP CONSTRAINT IF EXISTS books_status_check;
ALTER TABLE books ADD CONSTRAINT books_status_check CHECK (status IN ('na_estante', 'quero_ler', 'lendo', 'lido'));
