CREATE TABLE reading_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    book_id UUID NOT NULL REFERENCES books (id) ON DELETE CASCADE,
    log_date DATE NOT NULL,
    pages_read INTEGER NOT NULL CHECK (pages_read > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT reading_logs_unique_book_per_day UNIQUE (book_id, log_date)
);

CREATE INDEX idx_reading_logs_user_id_log_date ON reading_logs (user_id, log_date);
CREATE INDEX idx_reading_logs_book_id ON reading_logs (book_id);
