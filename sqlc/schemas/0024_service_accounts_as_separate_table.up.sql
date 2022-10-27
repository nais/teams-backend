BEGIN;

CREATE TABLE service_accounts (
   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
   name TEXT NOT NULL UNIQUE
);

INSERT INTO service_accounts (id, name)
SELECT id, name
FROM users
WHERE service_account = true;

DELETE FROM users
WHERE service_account = true;

ALTER TABLE users
    DROP COLUMN service_account,
    ALTER COLUMN email SET NOT NULL,
    ALTER COLUMN external_id SET NOT NULL;

ALTER TABLE api_keys RENAME COLUMN user_id TO service_account_id;

ALTER TABLE api_keys
    DROP CONSTRAINT api_keys_user_id_fkey,
    ADD CONSTRAINT api_keys_service_account_id_fkey FOREIGN KEY(service_account_id) REFERENCES service_accounts(id) ON DELETE CASCADE;

COMMIT;