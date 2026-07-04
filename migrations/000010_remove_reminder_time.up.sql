ALTER TABLE push_subscriptions
    DROP COLUMN IF EXISTS reminder_hour,
    DROP COLUMN IF EXISTS reminder_minute;
