BEGIN;

ALTER TABLE system_states RENAME TO reconciler_states;

COMMIT;