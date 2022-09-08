BEGIN;

ALTER TABLE users DROP CONSTRAINT users_email_required_for_regular_users_check;
ALTER TABLE users DROP CONSTRAINT users_email_not_allowed_for_service_accounts_check;

/*
    The email address should preferably have the tenant domain but we don't have a way of injecting that value into
    the SQL yet, so we just generate a random UUID and append the statix suffix.
*/
UPDATE users SET email = name || '@' || gen_random_uuid() || '.serviceaccounts.nais.io' WHERE service_account = true;
ALTER TABLE users DROP COLUMN service_account;
ALTER TABLE users ALTER COLUMN email SET NOT NULL;

COMMIT;