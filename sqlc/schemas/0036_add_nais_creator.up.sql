BEGIN;

ALTER TYPE role_name ADD VALUE IF NOT EXISTS 'NaisTeam creator';
ALTER TYPE authz_name ADD VALUE IF NOT EXISTS 'teams:skip_nais_validation';

COMMIT;
BEGIN;
INSERT INTO role_authz(role_name, authz_name) VALUES
    ('NaisTeam creator','teams:skip_nais_validation'),
    ('NaisTeam creator','teams:create');

COMMIT;
