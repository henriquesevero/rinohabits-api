ALTER TABLE courses DROP COLUMN IF EXISTS sort_order;

ALTER TABLE courses DROP CONSTRAINT courses_status_check;
ALTER TABLE courses ADD CONSTRAINT courses_status_check
  CHECK (status IN ('quero_fazer', 'fazendo', 'concluido'));
