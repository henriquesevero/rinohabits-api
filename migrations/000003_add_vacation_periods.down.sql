DROP INDEX IF EXISTS idx_vacation_periods_one_open_per_user;
DROP INDEX IF EXISTS idx_vacation_periods_user_id;
DROP TABLE IF EXISTS vacation_periods;
ALTER TABLE users ADD COLUMN vacation_mode BOOLEAN NOT NULL DEFAULT FALSE;
