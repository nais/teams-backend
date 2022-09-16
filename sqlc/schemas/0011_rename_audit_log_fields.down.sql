BEGIN;

ALTER TABLE audit_logs RENAME actor TO actor_email;
ALTER TABLE audit_logs RENAME target_user TO target_user_email;

COMMIT;