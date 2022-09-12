BEGIN;

/* rename groupEmail key to groupId, keep the value */
UPDATE system_states
SET state = jsonb_build_object('groupId', state->'groupEmail') || state - 'groupEmail'
WHERE system_name = 'google:workspace-admin';

COMMIT;