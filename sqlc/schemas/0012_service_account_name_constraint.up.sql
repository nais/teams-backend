BEGIN;

ALTER TABLE users ADD CONSTRAINT users_service_account_name_check CHECK (name ~* '^[a-z][a-z0-9-]*[a-z0-9]$' OR service_account = false);

COMMIT;