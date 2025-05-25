-- +goose Up
-- +goose StatementBegin
ALTER TABLE reminders
    ADD COLUMN source_type VARCHAR(64) NOT NULL DEFAULT 'wishlist',
    ADD COLUMN source_id INTEGER,
    ADD COLUMN friend_name VARCHAR(64) NOT NULL DEFAULT '',
    ADD COLUMN friend_username VARCHAR(255) NOT NULL DEFAULT '',
    ADD COLUMN username VARCHAR(64) NOT NULL DEFAULT '';

CREATE INDEX idx_reminders_source ON reminders (source_type, source_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_reminders_source;

ALTER TABLE reminders
DROP COLUMN source_type,
DROP COLUMN source_id,
DROP COLUMN friend_name,
DROP COLUMN friend_username,
DROP COLUMN username;
-- +goose StatementEnd
