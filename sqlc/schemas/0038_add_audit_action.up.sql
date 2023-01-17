BEGIN;

ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'google:gcp:project:enable-google-apis';

COMMIT;