BEGIN;

/* Create tables and copy data */

CREATE TABLE service_accounts (
   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
   name TEXT NOT NULL UNIQUE
);

INSERT INTO service_accounts (id, name)
SELECT id, name
FROM users
WHERE service_account = true;

CREATE TABLE service_account_roles (
    id SERIAL PRIMARY KEY,
    role_name role_name NOT NULL,
    service_account_id UUID NOT NULL REFERENCES service_accounts(id) ON DELETE CASCADE,
    target_id UUID
);

INSERT INTO service_account_roles (role_name, service_account_id, target_id)
SELECT role_name, user_id, target_id FROM user_roles WHERE user_id IN (SELECT id FROM service_accounts);

/* Remove old data */

DELETE FROM users
WHERE service_account = true;

DELETE FROM user_roles
WHERE user_id IN (SELECT id FROM service_accounts);

/* Update existing schema */

ALTER TABLE users
    DROP COLUMN service_account,
    ALTER COLUMN email SET NOT NULL,
    ALTER COLUMN external_id SET NOT NULL;

ALTER TABLE api_keys RENAME COLUMN user_id TO service_account_id;
ALTER TABLE api_keys
    DROP CONSTRAINT api_keys_user_id_fkey,
    ADD CONSTRAINT api_keys_service_account_id_fkey FOREIGN KEY(service_account_id) REFERENCES service_accounts(id) ON DELETE CASCADE;

COMMIT;