BEGIN;

UPDATE teams SET name = slug WHERE TRIM(name) = '';
UPDATE teams SET purpose = 'NAIS team' WHERE TRIM(purpose) = '';
ALTER TABLE teams DROP CONSTRAINT teams_name_not_empty_check;
ALTER TABLE teams ADD CONSTRAINT teams_name_not_empty_check CHECK (TRIM(name) != '');
ALTER TABLE teams ADD CONSTRAINT teams_purpose_not_empty_check CHECK (TRIM(purpose) != '');

COMMIT;