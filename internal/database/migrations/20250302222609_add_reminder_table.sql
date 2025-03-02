-- +goose Up
-- +goose StatementBegin
CREATE TYPE reminder_status AS ENUM ('pending', 'scheduled');

CREATE TABLE reminders (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL references users(id),
    email VARCHAR(255) NOT NULL,
    title VARCHAR(255) NOT NULL,
    body VARCHAR(255) NOT NULL,
    type VARCHAR(32) NOT NULL,
    status reminder_status default 'pending',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE reminders;
DROP TYPE reminder_status;
-- +goose StatementEnd
