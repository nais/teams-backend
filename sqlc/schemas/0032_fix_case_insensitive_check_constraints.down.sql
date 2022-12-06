BEGIN;

ALTER TABLE teams DROP CONSTRAINT teams_slug_check;
ALTER TABLE teams ADD CONSTRAINT teams_slug_check CHECK (slug ~* '^(?=.{3,30}$)[a-z](-?[a-z0-9]+)+$');

COMMIT;
