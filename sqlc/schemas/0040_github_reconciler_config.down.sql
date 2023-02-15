BEGIN;

ALTER TYPE reconciler_config_key ADD VALUE IF NOT EXISTS 'github:app_id';
ALTER TYPE reconciler_config_key ADD VALUE IF NOT EXISTS 'github:app_private_key';

COMMIT;

BEGIN;

INSERT INTO reconciler_config (reconciler, key, display_name, description)
VALUES
    ('github:team', 'github:app_id',          'GitHub App ID',   'The application ID of the GitHub Application that Console will use when communicating with the GitHub APIs. The application will need the following permissions: Organization administration (read-only), Organization members (read and write).'),
    ('github:team', 'github:app_private_key', 'App private key', 'The private key of the GitHub Application (PEM format).')
;

COMMIT;
