BEGIN;

ALTER TYPE role_name ADD VALUE IF NOT EXISTS 'Deploy key viewer';
ALTER TYPE authz_name ADD VALUE IF NOT EXISTS 'deploy_key:view';

COMMIT;
BEGIN;
INSERT INTO role_authz(role_name, authz_name) VALUES
    ('Deploy key viewer','deploy_key:view'),
    ('Team member','deploy_key:view'),
    ('Team owner','deploy_key:view');

COMMIT;
