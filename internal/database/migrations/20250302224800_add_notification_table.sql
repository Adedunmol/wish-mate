-- +goose Up
-- +goose StatementBegin
CREATE TABLE notifications (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL references users(id),
    title VARCHAR(255) NOT NULL,
    body VARCHAR(255) NOT NULL,
    type VARCHAR(32) NOT NULL,
    read BOOLEAN default false,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE notifications;
-- +goose StatementEnd
