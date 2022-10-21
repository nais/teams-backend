BEGIN;

ALTER TABLE teams ADD COLUMN name TEXT UNIQUE;
UPDATE teams SET name = slug;
ALTER TABLE teams
    ALTER COLUMN name SET NOT NULL,
    ADD CONSTRAINT teams_name_not_empty_check CHECK (TRIM(name) != '');

COMMIT;