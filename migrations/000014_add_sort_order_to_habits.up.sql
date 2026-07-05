ALTER TABLE habits ADD COLUMN sort_order INTEGER NOT NULL DEFAULT 0;

WITH ranked AS (
  SELECT id,
    ROW_NUMBER() OVER (PARTITION BY user_id ORDER BY created_at ASC) - 1 AS rn
  FROM habits
  WHERE deleted_at IS NULL
)
UPDATE habits SET sort_order = ranked.rn FROM ranked WHERE habits.id = ranked.id;
