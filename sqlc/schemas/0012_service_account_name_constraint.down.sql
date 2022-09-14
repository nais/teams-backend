BEGIN;

ALTER TABLE users DROP CONSTRAINT users_service_account_name_check;

COMMIT;