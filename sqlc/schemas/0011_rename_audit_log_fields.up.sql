BEGIN;

ALTER TABLE audit_logs RENAME actor_email TO actor;
ALTER TABLE audit_logs RENAME target_user_email TO target_user;

COMMIT;