-- +goose Up
-- +goose StatementBegin
ALTER TABLE reminders
ADD COLUMN date DATE DEFAULT NOW();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE reminders
    DROP COLUMN date;
-- +goose StatementEnd
