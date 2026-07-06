ALTER TABLE courses DROP CONSTRAINT courses_status_check;
ALTER TABLE courses ADD CONSTRAINT courses_status_check
  CHECK (status IN ('quero_fazer', 'fazendo', 'concluido', 'na_prateleira'));

ALTER TABLE courses ADD COLUMN sort_order INT NOT NULL DEFAULT 0;

WITH ranked AS (
  SELECT id, ROW_NUMBER() OVER (PARTITION BY user_id ORDER BY created_at ASC) AS rn
  FROM courses
)
UPDATE courses SET sort_order = ranked.rn FROM ranked WHERE courses.id = ranked.id;
