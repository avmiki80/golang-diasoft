-- +goose Up
-- +goose StatementBegin

CREATE TABLE notifications (
                               id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                               title VARCHAR(255) NOT NULL,
                               created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
                               start_date TIMESTAMPTZ NOT NULL,
                               end_date TIMESTAMPTZ NOT NULL,
                               event_id UUID NOT NULL,
                               user_id UUID NOT NULL,

                               CONSTRAINT valid_notification_dates CHECK (end_date > start_date)
);

CREATE INDEX idx_notifications_event_id ON notifications(event_id);
CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_start_date ON notifications(start_date);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_notifications_start_date;
DROP INDEX IF EXISTS idx_notifications_user_id;
DROP INDEX IF EXISTS idx_notifications_event_id;
DROP TABLE IF EXISTS notifications;
-- +goose StatementEnd
