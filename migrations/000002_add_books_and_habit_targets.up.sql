ALTER TABLE habits
    ADD COLUMN monthly_target INTEGER CHECK (monthly_target IS NULL OR monthly_target > 0);

CREATE TABLE books (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    author TEXT,
    status TEXT NOT NULL DEFAULT 'quero_ler' CHECK (status IN ('quero_ler', 'lendo', 'lido')),
    total_pages INTEGER CHECK (total_pages IS NULL OR total_pages > 0),
    current_page INTEGER NOT NULL DEFAULT 0 CHECK (current_page >= 0),
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT books_current_page_within_total CHECK (
        total_pages IS NULL OR current_page <= total_pages
    )
);

CREATE INDEX idx_books_user_id ON books (user_id);
CREATE INDEX idx_books_user_id_status ON books (user_id, status);
CREATE INDEX idx_books_user_id_finished_at ON books (user_id, finished_at) WHERE finished_at IS NOT NULL;

CREATE TRIGGER books_set_updated_at
    BEFORE UPDATE ON books
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();
