BEGIN;

/* rename groupId key to groupEmail, keep the value */
UPDATE system_states
SET state = jsonb_build_object('groupEmail', state->'groupId') || state - 'groupId'
WHERE system_name = 'google:workspace-admin';

COMMIT;