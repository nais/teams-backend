BEGIN;

DROP INDEX audit_logs_target_team_slug_idx;
DROP INDEX audit_logs_created_at_desc_idx;

COMMIT;