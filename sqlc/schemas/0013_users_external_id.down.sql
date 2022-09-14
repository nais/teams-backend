BEGIN;

ALTER TABLE users DROP CONSTRAINT users_external_id_check;
ALTER TABLE users DROP COLUMN external_id;

COMMIT;