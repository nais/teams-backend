BEGIN;

/* Must use IF NOT EXISTS, since the value can potentially exist */
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'usersync:assign-admin-role';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'usersync:revoke-admin-role';

COMMIT;