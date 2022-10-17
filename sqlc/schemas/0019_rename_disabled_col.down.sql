BEGIN;

ALTER TABLE teams RENAME COLUMN enabled TO disabled;
ALTER TABLE teams ALTER COLUMN disabled SET DEFAULT false;
UPDATE teams SET disabled = NOT disabled;

COMMIT;