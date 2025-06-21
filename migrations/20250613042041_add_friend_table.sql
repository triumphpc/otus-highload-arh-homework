-- +goose Up
-- +goose StatementBegin
CREATE TABLE friends (
                         user_id INT NOT NULL,
                         friend_id INT NOT NULL,
                         created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
                         PRIMARY KEY (user_id, friend_id),
                         FOREIGN KEY (user_id) REFERENCES users(id),
                         FOREIGN KEY (friend_id) REFERENCES users(id),
                         CHECK (user_id <> friend_id)
);

CREATE INDEX idx_friends_friend_id_user_id ON friends(friend_id, user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Удаление дружбы между пользователями (удаляет обе связи)
DROP INDEX IF EXISTS idx_friends_friend_id_user_id;
DROP TABLE IF EXISTS friends;
-- +goose StatementEnd
