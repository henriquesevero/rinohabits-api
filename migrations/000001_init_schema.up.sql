CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE FUNCTION set_updated_at() RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    timezone TEXT NOT NULL DEFAULT 'UTC',
    monthly_reward_goal_cents INTEGER NOT NULL DEFAULT 0 CHECK (monthly_reward_goal_cents >= 0),
    vacation_mode BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TRIGGER users_set_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TABLE habits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    icon TEXT,
    color TEXT,
    active_weekdays SMALLINT[] NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT habits_active_weekdays_not_empty CHECK (array_length(active_weekdays, 1) >= 1),
    CONSTRAINT habits_active_weekdays_valid_range CHECK (
        active_weekdays <@ ARRAY[1, 2, 3, 4, 5, 6, 7]::SMALLINT[]
    )
);

CREATE INDEX idx_habits_user_id ON habits (user_id);
CREATE INDEX idx_habits_user_id_active ON habits (user_id) WHERE is_active AND deleted_at IS NULL;

CREATE TRIGGER habits_set_updated_at
    BEFORE UPDATE ON habits
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TABLE daily_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    habit_id UUID NOT NULL REFERENCES habits (id) ON DELETE CASCADE,
    log_date DATE NOT NULL,
    completed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT daily_logs_unique_habit_per_day UNIQUE (habit_id, log_date)
);

CREATE INDEX idx_daily_logs_user_id_log_date ON daily_logs (user_id, log_date);
CREATE INDEX idx_daily_logs_habit_id ON daily_logs (habit_id);

CREATE TABLE monthly_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    reference_month DATE NOT NULL,
    completion_percentage NUMERIC(5, 2) NOT NULL CHECK (
        completion_percentage >= 0 AND completion_percentage <= 100
    ),
    is_victory BOOLEAN NOT NULL,
    reward_goal_cents INTEGER NOT NULL CHECK (reward_goal_cents >= 0),
    consolidated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT monthly_history_unique_user_month UNIQUE (user_id, reference_month),
    CONSTRAINT monthly_history_reference_month_is_first_day CHECK (
        reference_month = date_trunc('month', reference_month)::DATE
    )
);

CREATE INDEX idx_monthly_history_user_id_reference_month ON monthly_history (user_id, reference_month);
