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
UPDATE reconciler_config SET key = 'github:org' WHERE reconciler = 'github:team' AND key = 'org';
UPDATE reconciler_config SET key = 'github:app_id' WHERE reconciler = 'github:team' AND key = 'app_id';
UPDATE reconciler_config SET key = 'github:app_installation_id' WHERE reconciler = 'github:team' AND key = 'app_installation_id';
UPDATE reconciler_config SET key = 'github:app_private_key' WHERE reconciler = 'github:team' AND key = 'app_private_key';

ALTER TABLE reconciler_config ALTER COLUMN key TYPE reconciler_config_key USING key::reconciler_config_key;

INSERT INTO reconciler_config (reconciler, key, description)
VALUES
    ('azure:group', 'azure:tenant_id', 'The ID of the Azure AD tenant.');

ALTER TABLE audit_logs DROP CONSTRAINT target_user_or_target_team;

COMMIT;