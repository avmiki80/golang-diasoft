-- +goose Up
-- +goose StatementBegin

-- Расширение для UUID
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE events (
                        id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                        title VARCHAR(255) NOT NULL,
                        start_date TIMESTAMPTZ NOT NULL,
                        end_date TIMESTAMPTZ NOT NULL,
                        description TEXT,
                        user_id UUID NOT NULL,
                        offset_time BIGINT DEFAULT 1,

                        CONSTRAINT valid_dates CHECK (end_date > start_date)

);

CREATE INDEX idx_events_user_id ON events(user_id);
CREATE INDEX idx_events_start_date ON events(start_date);
CREATE INDEX idx_events_end_date ON events(end_date);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_events_end_date;
DROP INDEX IF EXISTS idx_events_start_date;
DROP INDEX IF EXISTS idx_events_user_id;
DROP TABLE IF EXISTS events;
DROP EXTENSION IF EXISTS "uuid-ossp";
-- +goose StatementEnd