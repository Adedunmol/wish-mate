-- +goose Up
-- +goose StatementBegin
CREATE TYPE friendship_status AS ENUM ('pending', 'accepted', 'blocked');

CREATE TABLE friendships (
    status friendship_status NOT NULL default 'pending',
    user_id INTEGER NOT NULL references users(id),
    friend_id INTEGER NOT NULL references users(id),
    friend_since DATETIME,
    PRIMARY KEY (user_id, friend_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE friendships;
DROP TYPE friendship_status;
-- +goose StatementEnd
