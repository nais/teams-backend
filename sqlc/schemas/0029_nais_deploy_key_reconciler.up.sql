BEGIN;

ALTER TYPE reconciler_name ADD VALUE IF NOT EXISTS 'nais:deploy';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'nais:deploy:provision-deploy-key';

COMMIT;
BEGIN;

INSERT INTO reconcilers (name, display_name, description, run_order, enabled)
VALUES ('nais:deploy', 'NAIS deploy', 'Provision NAIS deploy key for Console teams.', 6, false);

COMMIT;