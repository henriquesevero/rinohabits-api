DROP INDEX idx_habits_user_id_active;

ALTER TABLE habits ADD COLUMN deleted_at TIMESTAMPTZ;

CREATE INDEX idx_habits_user_id_active ON habits (user_id) WHERE is_active AND deleted_at IS NULL;
