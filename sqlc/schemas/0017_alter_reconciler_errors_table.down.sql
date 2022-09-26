BEGIN;

ALTER TABLE reconciler_states RENAME TO system_states;

COMMIT;