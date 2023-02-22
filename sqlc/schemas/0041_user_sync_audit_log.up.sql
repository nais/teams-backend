BEGIN;

ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'graphql-api:users:sync';
ALTER TYPE audit_logs_target_type ADD VALUE IF NOT EXISTS 'system';

COMMIT;