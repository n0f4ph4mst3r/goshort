-- +goose Up
CREATE TABLE IF NOT EXISTS url (
    id SERIAL PRIMARY KEY,       
    alias TEXT NOT NULL UNIQUE,
    origin TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);

-- +goose Down
DROP TABLE IF EXISTS url;
