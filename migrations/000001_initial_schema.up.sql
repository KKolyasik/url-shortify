CREATE TABLE urls (
    id UUID PRIMARY KEY,
    original_url TEXT NOT NULL UNIQUE,
    short_code VARCHAR(22) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_urls_short_code ON urls (short_code);
