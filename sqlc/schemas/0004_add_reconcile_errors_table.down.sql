BEGIN;

DROP INDEX idx_reconcile_errors_created_at_desc;
DROP TABLE reconcile_errors;

COMMIT;