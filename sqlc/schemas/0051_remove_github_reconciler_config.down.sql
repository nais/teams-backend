BEGIN;

ALTER TYPE reconciler_config_key ADD VALUE IF NOT EXISTS 'github:org';
ALTER TYPE reconciler_config_key ADD VALUE IF NOT EXISTS 'github:app_installation_id';

COMMIT;

BEGIN;

INSERT INTO reconciler_config (reconciler, key, display_name, description)
VALUES
    ('github:team', 'github:org', 'Organization', 'The slug of the GitHub organization.'),
    ('github:team', 'github:app_installation_id', 'App installation ID', 'The installation ID for the GitHub application when installed on the org.')
;

COMMIT;
