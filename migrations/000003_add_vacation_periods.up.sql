ALTER TABLE users DROP COLUMN vacation_mode;

CREATE TABLE vacation_periods (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ended_at TIMESTAMPTZ,
    CONSTRAINT vacation_periods_ended_after_started CHECK (ended_at IS NULL OR ended_at > started_at)
);

CREATE INDEX idx_vacation_periods_user_id ON vacation_periods (user_id);
CREATE UNIQUE INDEX idx_vacation_periods_one_open_per_user ON vacation_periods (user_id) WHERE ended_at IS NULL;
