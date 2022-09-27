BEGIN;

ALTER TABLE reconciler_states RENAME CONSTRAINT reconciler_states_team_id_fkey TO system_states_team_id_fkey;
ALTER TABLE reconciler_states RENAME CONSTRAINT reconciler_states_pkey TO system_states_pkey;
ALTER TABLE reconciler_states DROP CONSTRAINT reconciler_states_reconciler_fkey;
ALTER TABLE reconciler_states ALTER COLUMN reconciler TYPE system_name USING reconciler::text::system_name;
ALTER TABLE reconciler_states RENAME COLUMN reconciler TO system_name;
ALTER TABLE reconciler_states RENAME TO system_states;

ALTER TABLE reconciler_errors RENAME CONSTRAINT reconciler_errors_team_id_fkey TO reconcile_errors_team_id_fkey;
ALTER INDEX reconciler_errors_team_id_reconciler_key RENAME TO reconcile_errors_team_id_system_name_key;
ALTER INDEX idx_reconciler_errors_created_at_desc RENAME TO idx_reconcile_errors_created_at_desc;
ALTER TABLE reconciler_errors RENAME CONSTRAINT reconciler_errors_pkey TO reconcile_errors_pkey;
ALTER TABLE reconciler_errors DROP CONSTRAINT reconciler_errors_reconciler_fkey;
ALTER TABLE reconciler_errors ALTER COLUMN reconciler TYPE system_name USING reconciler::text::system_name;
ALTER TABLE reconciler_errors RENAME COLUMN reconciler TO system_name;
ALTER TABLE reconciler_errors RENAME TO reconcile_errors;

ALTER TABLE reconciler_config ALTER COLUMN key TYPE TEXT;
DELETE FROM reconciler_config WHERE reconciler = 'azure:group' AND key = 'azure:tenant_id';
UPDATE reconciler_config SET key = 'client_id' WHERE reconciler = 'azure:group' AND key = 'azure:client_id';
UPDATE reconciler_config SET key = 'client_secret' WHERE reconciler = 'azure:group' AND key = 'azure:client_secret';
UPDATE reconciler_config SET key = 'org' WHERE reconciler = 'github:team' AND key = 'github:org';
UPDATE reconciler_config SET key = 'app_id' WHERE reconciler = 'github:team' AND key = 'github:app_id';
UPDATE reconciler_config SET key = 'app_installation_id' WHERE reconciler = 'github:team' AND key = 'github:app_installation_id';
UPDATE reconciler_config SET key = 'app_private_key' WHERE reconciler = 'github:team' AND key = 'github:app_private_key';
DROP TYPE reconciler_config_key;

DELETE FROM audit_logs WHERE target_user IS NULL AND target_team_slug IS NULL;
ALTER TABLE audit_logs ADD CONSTRAINT target_user_or_target_team CHECK (target_user IS NOT NULL OR target_team_slug IS NOT NULL);

COMMIT;