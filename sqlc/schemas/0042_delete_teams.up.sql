BEGIN;

ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'graphql-api:teams:request-delete';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'graphql-api:teams:delete';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'azure:group:delete';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'github:team:delete';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'google:workspace-admin:delete';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'google:gar:delete';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'google:gcp:delete-project';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'nais:namespace:delete-namespace';

CREATE TABLE team_delete_keys (
   key UUID PRIMARY KEY DEFAULT gen_random_uuid(),
   team_slug TEXT REFERENCES teams(slug) ON DELETE CASCADE,
   created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
   created_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
   confirmed_at TIMESTAMP WITH TIME ZONE
);

ALTER TABLE teams DROP COLUMN enabled;

COMMIT;
