-- +goose Up
-- +goose StatementBegin
CREATE TYPE status AS ENUM ('pending', 'accepted', 'blocked')

CREATE TABLE friendships (
    status status NOT NULL default 'pending',
    user_id INTEGER NOT NULL references users(id),
    friend_id INTEGER NOT NULL references users(id),
    friend_since DATETIME,
    PRIMARY KEY (user_id, friend_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE friendships;
DROP TYPE status;
-- +goose StatementEnd
