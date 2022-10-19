BEGIN;

UPDATE teams SET purpose = 'NAIS team' WHERE purpose IS NULL OR purpose = '';
ALTER TABLE teams ALTER COLUMN purpose SET NOT NULL;

COMMIT;