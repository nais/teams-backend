BEGIN;

ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'dependencytrack:group:create';

COMMIT;