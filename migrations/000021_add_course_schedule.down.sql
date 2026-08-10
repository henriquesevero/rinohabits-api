ALTER TABLE courses DROP CONSTRAINT IF EXISTS courses_schedule_requires_reminder;
ALTER TABLE courses DROP CONSTRAINT IF EXISTS courses_reminder_minute_valid;
ALTER TABLE courses DROP CONSTRAINT IF EXISTS courses_reminder_hour_valid;
ALTER TABLE courses DROP CONSTRAINT IF EXISTS courses_active_weekdays_valid_range;

ALTER TABLE courses
    DROP COLUMN IF EXISTS active_weekdays,
    DROP COLUMN IF EXISTS reminder_hour,
    DROP COLUMN IF EXISTS reminder_minute;
