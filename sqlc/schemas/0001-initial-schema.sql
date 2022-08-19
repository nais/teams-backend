-- enums

CREATE TYPE role_name AS ENUM (
    'Admin',
    'Service account creator',
    'Service account owner',
    'Team creator',
    'Team member',
    'Team owner',
    'Team viewer',
    'User admin',
    'User viewer'
);

CREATE TYPE authz_name AS ENUM (
    'audit_logs:read',
    'service_accounts:create',
    'service_accounts:delete',
    'service_accounts:list',
    'service_accounts:read',
    'service_accounts:update',
    'system_states:delete',
    'system_states:read',
    'system_states:update',
    'teams:create',
    'teams:delete',
    'teams:list',
    'teams:read',
    'teams:update',
    'users:list',
    'users:update'
);

CREATE TYPE audit_action AS ENUM (
	'console:api-key:create',
	'console:api-key:delete',
	'console:service-account:create',
	'console:service-account:delete',
	'console:service-account:update',
	'console:team:add-member',
	'console:team:add-owner',
	'console:team:create',
	'console:team:remove-member',
	'console:team:set-member-role',
	'console:team:sync',
	'console:team:update'
);

CREATE TYPE system_name AS ENUM (
    'console',
    'azure:group',
    'github:team',
    'google:gcp:project',
    'google:workspace-admin',
    'nais:namespace'
);

-- api_keys

CREATE TABLE api_keys (
    api_key TEXT PRIMARY KEY,
    user_id UUID NOT NULL
);

-- audit_logs

CREATE TABLE audit_logs (
    id UUID PRIMARY KEY,
    correlation_id UUID NOT NULL,
    actor_email TEXT,
    actor_system system_name,
    target_user_email TEXT,
    target_team_slug TEXT,
    action audit_action NOT NULL,
    message TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,

    CHECK (actor_email IS NOT NULL OR actor_system IS NOT NULL),
    CHECK (actor_email IS NOT NULL AND actor_system IS NOT NULL),
    CHECK (target_user_email IS NOT NULL OR target_team_slug IS NOT NULL)
);


-- role_authz

CREATE TABLE role_authz (
    authz_name authz_name,
    role_name role_name,

    PRIMARY KEY (authz_name, role_name)
);

INSERT INTO role_authz(role_name, authz_name) VALUES
    ('Admin','audit_logs.read'),
    ('Admin','service_accounts.create'),
    ('Admin','service_accounts.delete'),
    ('Admin','service_accounts.list'),
    ('Admin','service_accounts.read'),
    ('Admin','service_accounts.update'),
    ('Admin','system_states.delete'),
    ('Admin','system_states.read'),
    ('Admin','system_states.update'),
    ('Admin','teams.create'),
    ('Admin','teams.delete'),
    ('Admin','teams.list'),
    ('Admin','teams.read'),
    ('Admin','teams.update'),
    ('Admin','users.list'),
    ('Admin','users.update'),
    ('Service account creator','service_accounts.create'),
    ('Service account owner','service_accounts.delete'),
    ('Service account owner','service_accounts.read'),
    ('Service account owner','service_accounts.update'),
    ('Team creator','teams.create'),
    ('Team member','teams.read'),
    ('Team member','audit_logs.read'),
    ('Team owner','teams.delete'),
    ('Team owner','teams.read'),
    ('Team owner','teams.update'),
    ('Team owner','audit_logs.read'),
    ('Team viewer','teams.list'),
    ('Team viewer','teams.read'),
    ('Team viewer','audit_logs.read'),
    ('User admin','users.list'),
    ('User admin','users.update'),
    ('User viewer','users.list');

-- system_states

CREATE TABLE system_states (
    id UUID PRIMARY KEY,
    system_name system_name NOT NULL,
    team_id UUID NOT NULL REFERENCES teams(id),
    state jsonb DEFAULT '{}'::jsonb NOT NULL,

    UNIQUE (system_name, team_id)
);

-- team_metadata

CREATE TABLE team_metadata (
    id UUID PRIMARY KEY,
    team_id UUID NOT NULL REFERENCES teams(id),
    key TEXT NOT NULL,
    value TEXT,

    UNIQUE (team_id, key)
);

-- teams

CREATE TABLE teams (
    id UUID PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL UNIQUE,
    purpose TEXT,

    CHECK (slug ~* '^[a-z][a-z-]{1,18}[a-z]$')
);

-- user_roles

CREATE TABLE user_roles (
    id UUID PRIMARY KEY,
    role_name role_name NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id),
    target_id UUID
);

CREATE UNIQUE INDEX idx_unique_role_user_target ON user_roles (user_id, role_name, target_id)
WHERE target_id IS NOT NULL;

CREATE UNIQUE INDEX idx_unique_role_user ON user_roles (user_id, role_name)
WHERE target_id IS NULL;

-- user_teams

CREATE TABLE user_teams (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    team_id UUID NOT NULL REFERENCES teams(id),

    UNIQUE (user_id, team_id)
);

-- users

CREATE TABLE users (
    id UUID PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL
);