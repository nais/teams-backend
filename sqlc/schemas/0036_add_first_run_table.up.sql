BEGIN;

CREATE TABLE first_run (
   first_run BOOL NOT NULL
);

INSERT INTO first_run VALUES(true);

COMMIT;
