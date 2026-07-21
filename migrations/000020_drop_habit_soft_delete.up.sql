DELETE FROM habits WHERE deleted_at IS NOT NULL;

DROP INDEX idx_habits_user_id_active;

ALTER TABLE habits DROP COLUMN deleted_at;

CREATE INDEX idx_habits_user_id_active ON habits (user_id) WHERE is_active;
