BEGIN;

CREATE INDEX audit_logs_target_team_slug_idx ON audit_logs (target_team_slug);
CREATE INDEX audit_logs_created_at_desc_idx ON audit_logs (created_at DESC);

COMMIT;