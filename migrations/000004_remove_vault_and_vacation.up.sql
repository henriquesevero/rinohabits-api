ALTER TABLE users DROP COLUMN monthly_reward_goal_cents;

DROP TABLE IF EXISTS monthly_history;

DROP INDEX IF EXISTS idx_vacation_periods_one_open_per_user;
DROP INDEX IF EXISTS idx_vacation_periods_user_id;
DROP TABLE IF EXISTS vacation_periods;
