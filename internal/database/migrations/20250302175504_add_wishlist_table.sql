-- +goose Up
-- +goose StatementBegin
CREATE TABLE wishlists (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    notify_before SMALLINT NOT NULL,
    date DATE NOT NULL,
    created_by SERIAL references users(id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE wishlists;
-- +goose StatementEnd
