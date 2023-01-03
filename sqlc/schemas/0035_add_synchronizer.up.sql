BEGIN;

ALTER TYPE role_name ADD VALUE IF NOT EXISTS 'Synchronizer';
ALTER TYPE authz_name ADD VALUE IF NOT EXISTS 'teams:synchronize';
ALTER TYPE authz_name ADD VALUE IF NOT EXISTS 'usersync:synchronize';

COMMIT;
BEGIN;
INSERT INTO role_authz(role_name, authz_name) VALUES
    ('Team owner','teams:synchronize'),
    ('Synchronizer','teams:synchronize'),
    ('Synchronizer','usersync:synchronize'),
    ('Admin','teams:synchronize'),
    ('Admin','usersync:synchronize');

COMMIT;
