BEGIN;

ALTER TABLE system_states RENAME TO reconciler_states;
ALTER TABLE reconciler_states RENAME COLUMN system_name TO reconciler;
ALTER TABLE reconciler_states ALTER COLUMN reconciler TYPE reconciler_name USING reconciler::text::reconciler_name;
ALTER TABLE reconciler_states ADD CONSTRAINT reconciler_states_reconciler_fkey FOREIGN KEY (reconciler) REFERENCES reconcilers(name) ON DELETE CASCADE;
ALTER TABLE reconciler_states RENAME CONSTRAINT system_states_pkey TO reconciler_states_pkey;
ALTER TABLE reconciler_states RENAME CONSTRAINT system_states_team_id_fkey TO reconciler_states_team_id_fkey;

ALTER TABLE reconcile_errors RENAME TO reconciler_errors;
ALTER TABLE reconciler_errors RENAME COLUMN system_name TO reconciler;
ALTER TABLE reconciler_errors ALTER COLUMN reconciler TYPE reconciler_name USING reconciler::text::reconciler_name;
ALTER TABLE reconciler_errors ADD CONSTRAINT reconciler_errors_reconciler_fkey FOREIGN KEY (reconciler) REFERENCES reconcilers(name) ON DELETE CASCADE;
ALTER TABLE reconciler_errors RENAME CONSTRAINT reconcile_errors_pkey TO reconciler_errors_pkey;
ALTER INDEX idx_reconcile_errors_created_at_desc RENAME TO idx_reconciler_errors_created_at_desc;
ALTER INDEX reconcile_errors_team_id_system_name_key RENAME TO reconciler_errors_team_id_reconciler_key;
ALTER TABLE reconciler_errors RENAME CONSTRAINT reconcile_errors_team_id_fkey TO reconciler_errors_team_id_fkey;

CREATE TYPE reconciler_config_key AS ENUM (
    'azure:client_id',
    'azure:client_secret',
    'azure:tenant_id',
    'github:org',
    'github:app_id',
    'github:app_installation_id',
    'github:app_private_key'
);

UPDATE reconciler_config SET key = 'azure:client_id' WHERE reconciler = 'azure:group' AND key = 'client_id';
UPDATE reconciler_config SET key = 'azure:client_secret' WHERE reconciler = 'azure:group' AND key = 'client_secret';
UPDATE reconciler_config SET key = 'azure:tenant_id' WHERE reconciler = 'azure:group' AND key = 'tenant_id';
UPDATE reconciler_config SET key = 'github:org' WHERE reconciler = 'github:team' AND key = 'org';
UPDATE reconciler_config SET key = 'github:app_id' WHERE reconciler = 'github:team' AND key = 'app_id';
UPDATE reconciler_config SET key = 'github:app_installation_id' WHERE reconciler = 'github:team' AND key = 'app_installation_id';
UPDATE reconciler_config SET key = 'github:app_private_key' WHERE reconciler = 'github:team' AND key = 'app_private_key';

ALTER TABLE reconciler_config ALTER COLUMN key TYPE reconciler_config_key USING key::reconciler_config_key;

CREATE TYPE audit_logs_target_type AS ENUM (
     'user',
    'team',
    'service_account',
    'reconciler'
);

ALTER TABLE audit_logs DROP CONSTRAINT target_user_or_target_team;

ALTER TABLE audit_logs
    ADD COLUMN target_type audit_logs_target_type,
    ADD COLUMN target_identifier TEXT;

/* Insert extra rows for the teams in the rows with both user and teams present */
INSERT INTO audit_logs (created_at, correlation_id, system_name, actor, action, message, target_type, target_identifier)
    SELECT created_at, correlation_id, system_name, actor, action, message, 'team', target_team_slug
    FROM audit_logs
    WHERE target_user IS NOT NULL AND target_team_slug IS NOT NULL;

UPDATE audit_logs
SET target_type = 'user', target_identifier = target_user
WHERE target_user IS NOT NULL;

UPDATE audit_logs
SET target_type = 'team', target_identifier = target_team_slug
WHERE target_team_slug IS NOT NULL AND target_user IS NULL;

ALTER TABLE audit_logs
    DROP COLUMN target_user,
    DROP COLUMN target_team_slug,
    ALTER COLUMN target_type SET NOT NULL,
    ALTER COLUMN target_identifier SET NOT NULL;

COMMIT;