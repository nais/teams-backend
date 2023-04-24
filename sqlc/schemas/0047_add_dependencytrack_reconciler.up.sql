BEGIN;

ALTER TYPE reconciler_name ADD VALUE IF NOT EXISTS 'nais:dependencytrack';

COMMIT;
BEGIN;

INSERT INTO reconcilers (name, display_name, description, run_order, enabled)
VALUES ('nais:dependencytrack', 'DependencyTrack', 'Create teams and users in dependencytrack', 8, false);

COMMIT;
