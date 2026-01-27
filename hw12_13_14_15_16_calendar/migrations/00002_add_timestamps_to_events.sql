-- +goose Up
-- +goose StatementBegin

CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

ALTER TABLE events
    ADD COLUMN created_at TIMESTAMPTZ,
ADD COLUMN updated_at TIMESTAMPTZ;

UPDATE events
SET
    created_at = COALESCE(start_date, CURRENT_TIMESTAMP),
    updated_at = COALESCE(start_date, CURRENT_TIMESTAMP)
WHERE created_at IS NULL OR updated_at IS NULL;

ALTER TABLE events
    ALTER COLUMN created_at SET NOT NULL,
ALTER COLUMN updated_at SET NOT NULL;

ALTER TABLE events
    ALTER COLUMN created_at SET DEFAULT CURRENT_TIMESTAMP,
ALTER COLUMN updated_at SET DEFAULT CURRENT_TIMESTAMP;

CREATE TRIGGER update_events_updated_at
    BEFORE UPDATE ON events
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_timestamp();

CREATE INDEX idx_events_created_at ON events(created_at DESC);
CREATE INDEX idx_events_updated_at ON events(updated_at DESC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TRIGGER IF EXISTS update_events_updated_at ON events;

DROP INDEX IF EXISTS idx_events_updated_at;
DROP INDEX IF EXISTS idx_events_created_at;

ALTER TABLE events
DROP COLUMN IF EXISTS updated_at,
DROP COLUMN IF EXISTS created_at;

DROP FUNCTION IF EXISTS trigger_set_timestamp;

-- +goose StatementEnd