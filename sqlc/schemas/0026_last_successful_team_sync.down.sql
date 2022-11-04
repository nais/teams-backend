BEGIN;

ALTER TABLE teams DROP COLUMN last_successful_sync;

COMMIT;