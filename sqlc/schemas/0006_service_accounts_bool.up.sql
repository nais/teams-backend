BEGIN;

ALTER TABLE users ADD COLUMN service_account bool NOT NULL DEFAULT false;
CREATE UNIQUE INDEX users_unique_service_account_name_idx ON users (name) WHERE service_account = true;
ALTER TABLE users ALTER COLUMN email DROP NOT NULL;
UPDATE users SET service_account = true, email = NULL WHERE email LIKE '%serviceaccounts.nais.io';
ALTER TABLE users ADD CONSTRAINT users_email_required_for_regular_users_check CHECK (email IS NOT NULL OR service_account = true);
ALTER TABLE users ADD CONSTRAINT users_email_not_allowed_for_service_accounts_check CHECK (email IS NULL OR service_account = false);

COMMIT;