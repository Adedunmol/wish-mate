-- +goose Up
-- +goose StatementBegin
ALTER TABLE reminders ADD COLUMN template VARCHAR(64);
ALTER TABLE reminders ADD COLUMN execute_at TIMESTAMP DEFAULT NOW();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE reminders DROP COLUMN template;
ALTER TABLE reminders DROP COLUMN execute_at;
-- +goose StatementEnd
