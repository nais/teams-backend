BEGIN;

ALTER TABLE teams DROP CONSTRAINT teams_name_not_empty_check;
ALTER TABLE teams DROP CONSTRAINT teams_purpose_not_empty_check;
ALTER TABLE teams ADD CONSTRAINT teams_name_not_empty_check CHECK (LENGTH(name) > 0);

COMMIT;