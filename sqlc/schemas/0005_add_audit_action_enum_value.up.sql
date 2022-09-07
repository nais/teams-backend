BEGIN;

/* Must use IF NOT EXISTS, since the value can potentially exist */
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'graphql-api:roles:assign-global-role' AFTER 'graphql-api:team:update';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'graphql-api:roles:revoke-global-role' AFTER 'graphql-api:roles:assign-global-role';

COMMIT;