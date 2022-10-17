BEGIN;

ALTER TABLE teams RENAME COLUMN disabled TO enabled;
ALTER TABLE teams ALTER COLUMN enabled SET DEFAULT true;
UPDATE teams SET enabled = NOT enabled;

COMMIT;