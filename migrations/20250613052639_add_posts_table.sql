-- +goose Up
-- +goose StatementBegin
-- Создание таблицы постов
CREATE TABLE posts (
                       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                       author_id INTEGER NOT NULL REFERENCES users(id),
                       text TEXT NOT NULL,
                       created_at TIMESTAMP NOT NULL DEFAULT NOW(),
                       updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Индексы для быстрого поиска
CREATE INDEX idx_posts_author_id ON posts(author_id);
CREATE INDEX idx_posts_created_at ON posts(created_at);

-- Триггер для обновления updated_at
CREATE OR REPLACE FUNCTION update_post_updated_at()
    RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_post_updated_at
    BEFORE UPDATE ON posts
    FOR EACH ROW
EXECUTE FUNCTION update_post_updated_at();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS posts CASCADE;
-- +goose StatementEnd
