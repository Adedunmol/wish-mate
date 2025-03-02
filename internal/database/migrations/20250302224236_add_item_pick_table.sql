-- +goose Up
-- +goose StatementBegin
CREATE TABLE item_picks (
    id SERIAL PRIMARY KEY,
    item_id INTEGER NOT NULL references items(id),
    user_id INTEGER NOT NULL references users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE item_picks;
-- +goose StatementEnd
