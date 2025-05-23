-- +goose Up
-- +goose StatementBegin
ALTER TABLE reminders
ADD COLUMN source_type VARCHAR(64) NOT NULL DEFAULT 'wishlist',
ADD COLUMN source_id INTEGER;

CREATE INDEX idx_reminders_source ON reminders (source_type, source_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_reminders_source;

ALTER TABLE reminders
DROP COLUMN source_type,
DROP COLUMN source_id;
-- +goose StatementEnd
