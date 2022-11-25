BEGIN;

/* Must use IF NOT EXISTS, since the value can potentially exist */
ALTER TYPE system_name ADD VALUE IF NOT EXISTS 'authn';

COMMIT;
