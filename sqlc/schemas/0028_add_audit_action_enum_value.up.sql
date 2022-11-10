BEGIN;

ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'graphql-api:reconcilers:update-team-state';

COMMIT;