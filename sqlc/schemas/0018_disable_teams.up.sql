BEGIN;

ALTER TABLE teams ADD COLUMN disabled BOOL NOT NULL DEFAULT false;
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'graphql-api:team:disable' AFTER 'graphql-api:team:create';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'graphql-api:team:enable' AFTER 'graphql-api:team:disable';

COMMIT;