BEGIN;

ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'google:gcp:project:delete-cnrm-service-account';

COMMIT;