BEGIN;

-- types

CREATE TYPE audit_action AS ENUM (
    'azure:group:add-member',
    'azure:group:add-members',
    'azure:group:create',
    'azure:group:delete',
    'azure:group:delete-member',
    'dependencytrack:group:create',
    'github:team:add-member',
    'github:team:add-members',
    'github:team:create',
    'github:team:delete',
    'github:team:delete-member',
    'github:team:map-sso-user',
    'google:gar:delete',
    'google:gcp:delete-project',
    'google:gcp:project:assign-permissions',
    'google:gcp:project:create-cnrm-service-account',
    'google:gcp:project:create-project',
    'google:gcp:project:delete-cnrm-service-account',
    'google:gcp:project:enable-google-apis',
    'google:gcp:project:set-billing-info',
    'google:workspace-admin:add-member',
    'google:workspace-admin:add-members',
    'google:workspace-admin:add-to-gke-security-group',
    'google:workspace-admin:create',
    'google:workspace-admin:delete',
    'google:workspace-admin:delete-member',
    'graphql-api:api-key:create',
    'graphql-api:api-key:delete',
    'graphql-api:reconcilers:configure',
    'graphql-api:reconcilers:disable',
    'graphql-api:reconcilers:enable',
    'graphql-api:reconcilers:reset',
    'graphql-api:reconcilers:update-team-state',
    'graphql-api:roles:assign-global-role',
    'graphql-api:roles:revoke-global-role',
    'graphql-api:service-account:create',
    'graphql-api:service-account:delete',
    'graphql-api:service-account:update',
    'graphql-api:team:add-member',
    'graphql-api:team:add-owner',
    'graphql-api:team:create',
    'graphql-api:team:disable',
    'graphql-api:team:enable',
    'graphql-api:team:remove-member',
    'graphql-api:team:set-member-role',
    'graphql-api:team:sync',
    'graphql-api:team:update',
    'graphql-api:teams:delete',
    'graphql-api:teams:request-delete',
    'graphql-api:users:sync',
    'legacy-importer:team:add-member',
    'legacy-importer:team:add-owner',
    'legacy-importer:team:create',
    'legacy-importer:user:create',
    'nais:deploy:provision-deploy-key',
    'nais:namespace:create-namespace',
    'nais:namespace:delete-namespace',
    'usersync:assign-admin-role',
    'usersync:create',
    'usersync:delete',
    'usersync:list:local',
    'usersync:list:remote',
    'usersync:prepare',
    'usersync:revoke-admin-role',
    'usersync:update'
);

CREATE TYPE audit_logs_target_type AS ENUM (
    'reconciler',
    'service_account',
    'system',
    'team',
    'user'
);

CREATE TYPE reconciler_config_key AS ENUM (
    'azure:client_id',
    'azure:client_secret',
    'azure:tenant_id',
    'github:app_installation_id',
    'github:org'
);

CREATE TYPE reconciler_name AS ENUM (
    'azure:group',
    'github:team',
    'google:gcp:gar',
    'google:gcp:project',
    'google:workspace-admin',
    'nais:dependencytrack',
    'nais:deploy',
    'nais:namespace'
);

CREATE TYPE role_name AS ENUM (
    'Admin',
    'Deploy key viewer',
    'Service account creator',
    'Service account owner',
    'Synchronizer',
    'Team creator',
    'Team member',
    'Team owner',
    'Team viewer',
    'User admin',
    'User viewer'
);

CREATE TYPE system_name AS ENUM (
    'authn',
    'azure:group',
    'console',
    'github:team',
    'google:gcp:gar',
    'google:gcp:project',
    'google:workspace-admin',
    'graphql-api',
    'legacy-importer',
    'nais:deploy',
    'nais:dependencytrack',
    'nais:namespace',
    'usersync'
);

-- tables

CREATE TABLE api_keys (
    api_key text NOT NULL,
    service_account_id uuid NOT NULL,
    PRIMARY KEY(api_key)
);

CREATE TABLE audit_logs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    correlation_id uuid NOT NULL,
    system_name system_name NOT NULL,
    actor text,
    action audit_action NOT NULL,
    message text NOT NULL,
    target_type audit_logs_target_type NOT NULL,
    target_identifier text NOT NULL,
    PRIMARY KEY(id)
);

CREATE TABLE first_run (
    first_run boolean NOT NULL
);

CREATE TABLE reconciler_errors (
    id BIGSERIAL,
    correlation_id uuid NOT NULL,
    reconciler reconciler_name NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    error_message text NOT NULL,
    team_slug text NOT NULL,
    PRIMARY KEY(id),
    UNIQUE (team_slug, reconciler)
);

CREATE TABLE reconciler_config (
    reconciler reconciler_name NOT NULL,
    key reconciler_config_key NOT NULL,
    display_name text NOT NULL,
    description text NOT NULL,
    value text,
    secret boolean DEFAULT true NOT NULL,
    PRIMARY KEY (reconciler, key)
);

CREATE TABLE reconciler_states (
    reconciler reconciler_name NOT NULL,
    state jsonb DEFAULT '{}'::jsonb NOT NULL,
    team_slug text NOT NULL,
    PRIMARY KEY (reconciler, team_slug)
);

CREATE TABLE reconcilers (
    name reconciler_name NOT NULL,
    display_name text NOT NULL,
    description text NOT NULL,
    enabled boolean DEFAULT false NOT NULL,
    run_order integer NOT NULL,
    PRIMARY KEY(name),
    UNIQUE(display_name),
    UNIQUE(run_order)
);

CREATE TABLE service_account_roles (
    id SERIAL,
    role_name role_name NOT NULL,
    service_account_id uuid NOT NULL,
    target_team_slug text,
    target_service_account_id uuid,
    PRIMARY KEY(id),
    CHECK (((target_team_slug IS NULL) OR (target_service_account_id IS NULL)))
);

CREATE TABLE service_accounts (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name text NOT NULL,
    PRIMARY KEY(id),
    UNIQUE(name)
);

CREATE TABLE sessions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    expires timestamp with time zone NOT NULL,
    PRIMARY KEY(id)
);

CREATE TABLE slack_alerts_channels (
    team_slug text NOT NULL,
    environment text NOT NULL,
    channel_name text NOT NULL,
    PRIMARY KEY (team_slug, environment),
    CHECK ((channel_name ~ '^#[a-z0-9æøå_-]{2,80}$'::text))
);

CREATE TABLE team_delete_keys (
    key uuid DEFAULT gen_random_uuid() NOT NULL,
    team_slug text,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    created_by uuid NOT NULL,
    confirmed_at timestamp with time zone,
    PRIMARY KEY(key)
);

CREATE TABLE teams (
    slug text NOT NULL,
    purpose text NOT NULL,
    last_successful_sync timestamp without time zone,
    slack_channel text NOT NULL,
    PRIMARY KEY(slug),
    CHECK ((TRIM(BOTH FROM purpose) <> ''::text)),
    CHECK ((slack_channel ~ '^#[a-z0-9æøå_-]{2,80}$'::text)),
    CHECK ((slug ~ '^(?=.{3,30}$)[a-z](-?[a-z0-9]+)+$'::text))
);

CREATE TABLE user_roles (
    id SERIAL,
    role_name role_name NOT NULL,
    user_id uuid NOT NULL,
    target_team_slug text,
    target_service_account_id uuid,
    PRIMARY KEY(id),
    CHECK (((target_team_slug IS NULL) OR (target_service_account_id IS NULL)))
);

CREATE TABLE users (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    email text NOT NULL,
    name text NOT NULL,
    external_id text NOT NULL,
    PRIMARY KEY(id),
    UNIQUE(email),
    UNIQUE(external_id)
);

-- additional indexes

CREATE INDEX ON audit_logs USING btree (created_at DESC);
CREATE INDEX ON reconciler_errors USING btree (created_at DESC);
CREATE UNIQUE INDEX ON service_account_roles USING btree (service_account_id, role_name) WHERE ((target_team_slug IS NULL) AND (target_service_account_id IS NULL));
CREATE UNIQUE INDEX ON user_roles USING btree (user_id, role_name) WHERE ((target_team_slug IS NULL) AND (target_service_account_id IS NULL));
CREATE UNIQUE INDEX ON service_account_roles USING btree (service_account_id, role_name, target_service_account_id) WHERE (target_service_account_id IS NOT NULL);
CREATE UNIQUE INDEX ON user_roles USING btree (user_id, role_name, target_service_account_id) WHERE (target_service_account_id IS NOT NULL);
CREATE UNIQUE INDEX ON service_account_roles USING btree (service_account_id, role_name, target_team_slug) WHERE (target_team_slug IS NOT NULL);
CREATE UNIQUE INDEX ON user_roles USING btree (user_id, role_name, target_team_slug) WHERE (target_team_slug IS NOT NULL);

-- foreign keys

ALTER TABLE api_keys
ADD FOREIGN KEY (service_account_id) REFERENCES service_accounts(id) ON DELETE CASCADE;

ALTER TABLE reconciler_config
ADD FOREIGN KEY (reconciler) REFERENCES reconcilers(name) ON DELETE CASCADE;

ALTER TABLE reconciler_errors
ADD FOREIGN KEY (reconciler) REFERENCES reconcilers(name) ON DELETE CASCADE,
ADD FOREIGN KEY (team_slug) REFERENCES teams(slug) ON DELETE CASCADE;

ALTER TABLE reconciler_states
ADD FOREIGN KEY (reconciler) REFERENCES reconcilers(name) ON DELETE CASCADE,
ADD FOREIGN KEY (team_slug) REFERENCES teams(slug) ON DELETE CASCADE;

ALTER TABLE service_account_roles
ADD FOREIGN KEY (service_account_id) REFERENCES service_accounts(id) ON DELETE CASCADE,
ADD FOREIGN KEY (target_service_account_id) REFERENCES service_accounts(id) ON DELETE CASCADE,
ADD FOREIGN KEY (target_team_slug) REFERENCES teams(slug) ON DELETE CASCADE;

ALTER TABLE sessions
ADD FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE team_delete_keys
ADD FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE,
ADD FOREIGN KEY (team_slug) REFERENCES teams(slug) ON DELETE CASCADE;

ALTER TABLE user_roles
ADD FOREIGN KEY (target_service_account_id) REFERENCES service_accounts(id) ON DELETE CASCADE,
ADD FOREIGN KEY (target_team_slug) REFERENCES teams(slug) ON DELETE CASCADE,
ADD FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- data

INSERT INTO first_run VALUES(true);

INSERT INTO reconcilers
(name, display_name, description, enabled, run_order) VALUES
('github:team', 'GitHub teams', 'Create and maintain GitHub teams for the Console teams.', false, 1),
('azure:group', 'Azure AD groups', 'Create and maintain Azure AD security groups for the Console teams.', false, 2),
('google:workspace-admin', 'Google workspace group', 'Create and maintain Google workspace groups for the Console teams.', false, 3),
('google:gcp:project', 'GCP projects', 'Create GCP projects for the Console teams.', false, 4),
('nais:namespace', 'NAIS namespace', 'Create NAIS namespaces for the Console teams.', false, 5),
('nais:deploy', 'NAIS deploy', 'Provision NAIS deploy key for Console teams.', false, 6),
('google:gcp:gar', 'Google Artifact Registry', 'Provision artifact registry repositories for Console teams.', false, 7),
('nais:dependencytrack', 'DependencyTrack', 'Create teams and users in dependencytrack', 8, false);

INSERT INTO reconciler_config
(reconciler, key, display_name, description, secret) VALUES
('azure:group', 'azure:client_secret', 'Client secret', 'The client secret of the application registration.', true),
('azure:group', 'azure:client_id', 'Client ID', 'The client ID of the application registration that Console will use when communicating with the Azure AD APIs. The application must have the following API permissions: Group.Create, GroupMember.ReadWrite.All.', false),
('azure:group', 'azure:tenant_id', 'Tenant ID', 'The ID of the Azure AD tenant.', false),
('github:team', 'github:org', 'Organization', 'The slug of the GitHub organization.', false),
('github:team', 'github:app_installation_id', 'App installation ID', 'The installation ID for the GitHub application when installed on the org.', false);

COMMIT;