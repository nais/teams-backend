BEGIN;

ALTER TYPE reconciler_name ADD VALUE IF NOT EXISTS 'google:gcp:gar';

COMMIT;
BEGIN;

INSERT INTO reconcilers (name, display_name, description, run_order, enabled)
VALUES ('google:gcp:gar', 'Google Artifact Registry', 'Provision artifact registry repositories for Console teams.', 7, true);

COMMIT;
