BEGIN;

/*
    The slug must:

    - contain only lowercase alphanumeric characters or hyphens
    - contain at least 3 characters and at most 30 characters
    - start with an alphabetic character
    - end with an alphanumeric character
    - not contain two hyphens in a row
*/

ALTER TABLE teams DROP CONSTRAINT teams_slug_check;
ALTER TABLE teams ADD CONSTRAINT teams_slug_check CHECK (slug ~* '^(?=.{3,30}$)[a-z](-?[a-z0-9]+)+$');

COMMIT;