-- +goose Up
-- +goose StatementBegin
DROP TABLE IF EXISTS users CASCADE;

CREATE TABLE IF EXISTS  users (
                       id BIGSERIAL PRIMARY KEY,
                       first_name TEXT NOT NULL,
                       last_name TEXT NOT NULL,
                       email TEXT NOT NULL,
                       birth_date DATE NOT NULL,
                       gender TEXT NOT NULL CHECK (gender IN ('male', 'female', 'other')),
                       interests TEXT[],
                       city TEXT NOT NULL,
                       password_hash TEXT NOT NULL,
                       created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                       updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

SELECT create_distributed_table('users', 'id');

CREATE TABLE IF EXISTS dialogs (
                         dialog_id BIGSERIAL,
                         user1_id BIGINT NOT NULL,
                         user2_id BIGINT NOT NULL,
                         created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                         PRIMARY KEY (dialog_id, user1_id),
                         CONSTRAINT user_order CHECK (user1_id < user2_id)
);

CREATE INDEX IF EXISTS  idx_dialogs_user_pair ON dialogs (user1_id, user2_id);

SELECT create_distributed_table('dialogs', 'user1_id', colocate_with => 'users');

CREATE TABLE IF EXISTS  messages (
                          message_id BIGSERIAL,
                          dialog_id BIGINT NOT NULL,
                          sender_id BIGINT NOT NULL,  -- Будем колоцировать с users.id
                          recipient_id BIGINT NOT NULL,
                          content TEXT NOT NULL,
                          created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                          read_at TIMESTAMPTZ,
                          PRIMARY KEY (message_id, sender_id)
);

SELECT create_distributed_table('messages', 'sender_id', colocate_with => 'users');


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
