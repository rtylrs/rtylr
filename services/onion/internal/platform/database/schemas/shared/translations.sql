CREATE TABLE translations (
    id SERIAL PRIMARY KEY,
    key_text VARCHAR(100) NOT NULL DEFAULT '',
    en TEXT,
    ar TEXT,
    fr TEXT,
    key_group VARCHAR(40) NOT NULL,
    auto_generated BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(key_text, key_group)
);

CREATE INDEX idx_translations_key_group ON translations(key_group);

CREATE INDEX idx_translations_key_text ON translations(key_text); 