BEGIN;

/* Must use IF NOT EXISTS, since the value can potentially exist */
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'google:gcp:project:set-billing-info' AFTER 'google:gcp:project:assign-permissions';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'google:gcp:project:create-cnrm-service-account' AFTER 'google:gcp:project:set-billing-info';

COMMIT;