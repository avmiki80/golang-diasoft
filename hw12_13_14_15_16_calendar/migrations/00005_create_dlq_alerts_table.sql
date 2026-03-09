-- +goose Up
-- +goose StatementBegin

CREATE TABLE dlq_alerts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    topic VARCHAR(255) NOT NULL,
    original_topic VARCHAR(255) NOT NULL,
    partition INT NOT NULL,
    message_offset BIGINT NOT NULL,
    message_key VARCHAR(255),
    error_message TEXT NOT NULL,
    error_count INT NOT NULL DEFAULT 1,
    original_message TEXT NOT NULL,
    failed_at TIMESTAMPTZ NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'new',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT valid_status CHECK (status IN ('new', 'processing', 'resolved', 'failed'))
);

CREATE INDEX idx_dlq_alerts_status ON dlq_alerts(status);
CREATE INDEX idx_dlq_alerts_original_topic ON dlq_alerts(original_topic);
CREATE INDEX idx_dlq_alerts_created_at ON dlq_alerts(created_at DESC);
CREATE INDEX idx_dlq_alerts_failed_at ON dlq_alerts(failed_at DESC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_dlq_alerts_failed_at;
DROP INDEX IF EXISTS idx_dlq_alerts_created_at;
DROP INDEX IF EXISTS idx_dlq_alerts_original_topic;
DROP INDEX IF EXISTS idx_dlq_alerts_status;
DROP TABLE IF EXISTS dlq_alerts;
-- +goose StatementEnd
