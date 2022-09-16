BEGIN;

ALTER TABLE users ADD COLUMN external_id TEXT;
UPDATE users SET external_id = email WHERE service_account = false;
CREATE UNIQUE INDEX users_external_id_idx ON users (external_id);
ALTER TABLE users ADD CONSTRAINT users_external_id_check CHECK (external_id IS NOT NULL OR service_account = true);

COMMIT;