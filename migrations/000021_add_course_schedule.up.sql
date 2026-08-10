ALTER TABLE courses
    ADD COLUMN active_weekdays SMALLINT[] NOT NULL DEFAULT '{}',
    ADD COLUMN reminder_hour SMALLINT,
    ADD COLUMN reminder_minute SMALLINT;

ALTER TABLE courses ADD CONSTRAINT courses_active_weekdays_valid_range CHECK (
    active_weekdays <@ ARRAY[1, 2, 3, 4, 5, 6, 7]::SMALLINT[]
);

ALTER TABLE courses ADD CONSTRAINT courses_reminder_hour_valid CHECK (
    reminder_hour IS NULL OR (reminder_hour BETWEEN 0 AND 23)
);

ALTER TABLE courses ADD CONSTRAINT courses_reminder_minute_valid CHECK (
    reminder_minute IS NULL OR (reminder_minute BETWEEN 0 AND 59)
);

ALTER TABLE courses ADD CONSTRAINT courses_schedule_requires_reminder CHECK (
    (array_length(active_weekdays, 1) IS NULL) = (reminder_hour IS NULL)
    AND (reminder_hour IS NULL) = (reminder_minute IS NULL)
);
