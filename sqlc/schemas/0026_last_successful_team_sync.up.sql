BEGIN;

ALTER TABLE teams ADD COLUMN last_successful_sync TIMESTAMP;

COMMIT;