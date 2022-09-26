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

COMMIT;