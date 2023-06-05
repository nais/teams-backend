BEGIN;

ALTER TABLE audit_logs ALTER COLUMN action TYPE TEXT;
ALTER TABLE audit_logs ALTER COLUMN system_name TYPE TEXT;
ALTER TABLE audit_logs ALTER COLUMN target_type TYPE TEXT;
ALTER TABLE audit_logs RENAME COLUMN system_name TO component_name;

DROP TYPE audit_action;
DROP TYPE system_name;
DROP TYPE audit_logs_target_type;

COMMIT;
