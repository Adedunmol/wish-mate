-- +goose Up
-- +goose StatementBegin
CREATE TYPE notification_status AS ENUM ('read', 'unread');

CREATE TABLE notifications (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL references users(id),
    title VARCHAR(255) NOT NULL,
    body VARCHAR(255) NOT NULL,
    type VARCHAR(32) NOT NULL,
    status notification_status default 'unread',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE notifications;
DROP TYPE notification_status;
-- +goose StatementEnd
