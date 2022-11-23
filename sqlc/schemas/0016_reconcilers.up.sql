BEGIN;

CREATE TYPE reconciler_name AS ENUM (
    'azure:group',
    'github:team',
    'google:gcp:project',
    'google:workspace-admin',
    'nais:namespace'
);

CREATE TABLE reconcilers (
    name reconciler_name PRIMARY KEY,
    display_name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL,
    enabled BOOL NOT NULL DEFAULT false,
    run_order INT NOT NULL UNIQUE
);

CREATE TABLE reconciler_config (
    reconciler reconciler_name NOT NULL REFERENCES reconcilers(name) ON DELETE CASCADE,
    key TEXT NOT NULL,
    display_name TEXT NOT NULL,
    description TEXT NOT NULL,
    value TEXT,
    PRIMARY KEY(reconciler, key)
);

INSERT INTO reconcilers (name, display_name, description, run_order, enabled)
VALUES
    ('google:workspace-admin',  'Google workspace group', 'Create and maintain Google workspace groups for the Console teams.',  1, false),
    ('google:gcp:project',      'GCP projects',           'Create GCP projects for the Console teams.',                          2, false),
    ('nais:namespace',          'NAIS namespace',         'Create NAIS namespaces for the Console teams.',                       3, false),
    ('azure:group',             'Azure AD groups',        'Create and maintain Azure AD security groups for the Console teams.', 4, false),
    ('github:team',             'GitHub teams',           'Create and maintain GitHub teams for the Console teams.',             5, false)
;

/* Add reconciler config that needs to be set by the tenant. Some reconciler options are provided via environment variables */
INSERT INTO reconciler_config (reconciler, key, display_name, description)
VALUES
    ('azure:group', 'azure:client_id',     'Client ID',      'The client ID of the application registration that Console will use when communicating with the Azure AD APIs. The application must have the following API permissions: Group.Create, GroupMember.ReadWrite.All.'),
    ('azure:group', 'azure:client_secret', 'Client secret',  'The client secret of the application registration.'),
    ('azure:group', 'azure:tenant_id',     'Tenant ID',      'The ID of the Azure AD tenant.'),

    ('github:team', 'github:org',                 'Organization',        'The slug of the GitHub organization.'),
    ('github:team', 'github:app_id',              'GitHub App ID',       'The application ID of the GitHub Application that Console will use when communicating with the GitHub APIs. The application will need the following permissions: Organization administration (read-only), Organization members (read and write).'),
    ('github:team', 'github:app_installation_id', 'App installation ID', 'The installation ID for the GitHub application when installed on the org.'),
    ('github:team', 'github:app_private_key',     'App private key',     'The private key of the GitHub Application (PEM format).')
;

/* Must use IF NOT EXISTS, since the value can potentially exist */
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'graphql-api:reconcilers:configure' AFTER 'graphql-api:roles:revoke-global-role';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'graphql-api:reconcilers:disable' AFTER 'graphql-api:reconcilers:configure';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'graphql-api:reconcilers:enable' AFTER 'graphql-api:reconcilers:disable';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'graphql-api:reconcilers:reset' AFTER 'graphql-api:reconcilers:enable';

COMMIT;