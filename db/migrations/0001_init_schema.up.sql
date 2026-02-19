CREATE TABLE IF NOT EXISTS users(
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nickname      VARCHAR(256) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS links (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code          VARCHAR(6) UNIQUE NOT NULL,
    original_url  TEXT NOT NULL,
    created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    user_id       UUID REFERENCES users(id) ON DELETE CASCADE
);