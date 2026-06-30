CREATE TABLE courses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT,
    link TEXT,
    status TEXT NOT NULL DEFAULT 'quero_fazer' CHECK (status IN ('quero_fazer', 'fazendo', 'concluido')),
    total_hours NUMERIC(7,2) CHECK (total_hours IS NULL OR total_hours > 0),
    current_hours NUMERIC(7,2) NOT NULL DEFAULT 0 CHECK (current_hours >= 0),
    cover_url TEXT,
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT courses_current_hours_within_total CHECK (
        total_hours IS NULL OR current_hours <= total_hours
    )
);

CREATE INDEX idx_courses_user_id ON courses (user_id);
CREATE INDEX idx_courses_user_id_status ON courses (user_id, status);

CREATE TRIGGER courses_set_updated_at
    BEFORE UPDATE ON courses
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TABLE course_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    course_id UUID NOT NULL REFERENCES courses (id) ON DELETE CASCADE,
    log_date DATE NOT NULL,
    hours_logged NUMERIC(6,2) NOT NULL CHECK (hours_logged > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT course_logs_unique_course_per_day UNIQUE (course_id, log_date)
);

CREATE INDEX idx_course_logs_user_id_log_date ON course_logs (user_id, log_date);
CREATE INDEX idx_course_logs_course_id ON course_logs (course_id);
