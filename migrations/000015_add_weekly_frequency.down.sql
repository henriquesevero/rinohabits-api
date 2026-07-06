ALTER TABLE habits DROP CONSTRAINT IF EXISTS habits_has_schedule;
ALTER TABLE habits DROP COLUMN IF EXISTS weekly_frequency;
ALTER TABLE habits ADD CONSTRAINT habits_active_weekdays_not_empty CHECK (array_length(active_weekdays, 1) >= 1);
