ALTER TABLE books ADD COLUMN sort_order INTEGER NOT NULL DEFAULT 0;

WITH ranked AS (
  SELECT id,
    ROW_NUMBER() OVER (
      PARTITION BY user_id
      ORDER BY
        CASE status
          WHEN 'na_estante' THEN 0
          WHEN 'quero_ler'  THEN 1
          WHEN 'lendo'      THEN 2
          WHEN 'lido'       THEN 3
        END,
        created_at DESC
    ) - 1 AS rn
  FROM books
)
UPDATE books SET sort_order = ranked.rn FROM ranked WHERE books.id = ranked.id;
