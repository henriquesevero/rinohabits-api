ALTER TABLE habits
    ADD COLUMN weekly_frequency SMALLINT
    CHECK (weekly_frequency IS NULL OR (weekly_frequency >= 1 AND weekly_frequency <= 7));

ALTER TABLE habits DROP CONSTRAINT habits_active_weekdays_not_empty;

ALTER TABLE habits ADD CONSTRAINT habits_has_schedule CHECK (
    array_length(active_weekdays, 1) >= 1 OR weekly_frequency IS NOT NULL
);
