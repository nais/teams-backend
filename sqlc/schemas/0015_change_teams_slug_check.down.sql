BEGIN;

ALTER TABLE teams DROP CONSTRAINT teams_slug_check;
ALTER TABLE teams ADD CONSTRAINT teams_slug_check CHECK (slug ~* '^[a-z][a-z-]{1,18}[a-z]$');

COMMIT;