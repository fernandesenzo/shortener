CREATE TABLE IF NOT EXISTS links (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code         VARCHAR(6) UNIQUE NOT NULL,
    original_url TEXT NOT NULL,
    created_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)