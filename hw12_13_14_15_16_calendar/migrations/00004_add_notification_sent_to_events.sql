-- +goose Up
-- +goose StatementBegin

ALTER TABLE events
    ADD COLUMN notification_sent BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX idx_events_notification_sent ON events(notification_sent);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_events_notification_sent;

ALTER TABLE events
    DROP COLUMN IF EXISTS notification_sent;

-- +goose StatementEnd
