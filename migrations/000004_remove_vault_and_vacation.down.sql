ALTER TABLE users ADD COLUMN monthly_reward_goal_cents INTEGER NOT NULL DEFAULT 0 CHECK (monthly_reward_goal_cents >= 0);

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

CREATE TABLE vacation_periods (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ended_at TIMESTAMPTZ,
    CONSTRAINT vacation_periods_ended_after_started CHECK (ended_at IS NULL OR ended_at > started_at)
);

CREATE INDEX idx_vacation_periods_user_id ON vacation_periods (user_id);
CREATE UNIQUE INDEX idx_vacation_periods_one_open_per_user ON vacation_periods (user_id) WHERE ended_at IS NULL;
