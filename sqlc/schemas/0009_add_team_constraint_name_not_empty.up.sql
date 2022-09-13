BEGIN;

-- Make sure that teams.name is a non-zero length string
UPDATE teams SET name = id WHERE LENGTH(name) = 0;
ALTER TABLE teams ADD CONSTRAINT teams_name_not_empty_check CHECK (LENGTH(name) > 0);

COMMIT;