BEGIN;

ALTER TABLE teams DROP CONSTRAINT teams_name_not_empty_check;

COMMIT;