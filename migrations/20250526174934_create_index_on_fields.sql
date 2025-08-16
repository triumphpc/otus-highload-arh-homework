-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX  IF NOT EXISTS users_name_trgm_idx ON users
USING gin (
    first_name gin_trgm_ops,
    last_name gin_trgm_ops
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS users_name_trgm_idx;

-- Удаляем расширение только если больше нет зависимостей
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_indexes
        WHERE indexdef LIKE '%gin_trgm_ops%'
    ) THEN
        DROP EXTENSION IF EXISTS pg_trgm;
    END IF;
END $$;
-- +goose StatementEnd
