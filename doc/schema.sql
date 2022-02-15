-- TODO:
-- - do we need child roles?
-- - do components have their own state (i.e., ID of created groups?)

CREATE TABLE users
(
    id                  uuid primary key           not null,
    email               varchar unique             null,
    apikey              varchar unique             null,
    name                varchar                    not null,
    deleted_by_user_id  uuid references users (id) null,
    created_by_user_id  uuid references users (id) null,
    modified_by_user_id uuid references users (id) null,
    deleted             timestamp with time zone   null,
    created             timestamp with time zone   not null,
    modified            timestamp with time zone   not null
);

CREATE TABLE teams
(
    id                  uuid primary key           not null,
    slug                varchar unique             not null,
    name                varchar unique             not null,
    purpose             varchar                    null,
    deleted_by_user_id  uuid references users (id) null,
    created_by_user_id  uuid references users (id) null,
    modified_by_user_id uuid references users (id) null,
    deleted             timestamp with time zone   null,
    created             timestamp with time zone   not null,
    modified            timestamp with time zone   not null
);

CREATE TABLE users_teams
(
    id                  uuid primary key           not null,
    user_id             uuid references users (id) not null,
    team_id             uuid references teams (id) not null,
    created_by_user_id  uuid references users (id) null,
    modified_by_user_id uuid references users (id) null,
    created             timestamp with time zone   not null,
    modified            timestamp with time zone   not null
);

CREATE UNIQUE INDEX users_teams_pk ON users_teams (user_id, team_id);

CREATE TABLE team_metadata
(
    id                  uuid primary key           not null,
    team_id             uuid                       not null,
    key                 varchar                    not null,
    value               varchar                    null,
    deleted_by_user_id  uuid references users (id) null,
    created_by_user_id  uuid references users (id) null,
    modified_by_user_id uuid references users (id) null,
    deleted             timestamp with time zone   null,
    created             timestamp with time zone   not null,
    modified            timestamp with time zone   not null
);

CREATE UNIQUE INDEX team_metadata_pk ON team_metadata (team_id, key);

CREATE TABLE systems
(
    id                  uuid primary key           not null,
    deleted_by_user_id  uuid references users (id) null,
    created_by_user_id  uuid references users (id) null,
    modified_by_user_id uuid references users (id) null,
    deleted             timestamp with time zone   null,
    created             timestamp with time zone   not null,
    modified            timestamp with time zone   not null
);

CREATE TABLE roles
(
    id                  uuid primary key             not null,
    system_id           uuid references systems (id) not null,
    resource            varchar                      not null, -- sub-resource at system (maybe not needed if systems are namespaced, e.g. gcp:buckets)
    access_level        varchar                      not null, -- read, write, R/W, other combinations per system
    permission          varchar                      not null, -- allow/deny
    deleted_by_user_id  uuid references users (id)   null,
    created_by_user_id  uuid references users (id)   null,
    modified_by_user_id uuid references users (id)   null,
    deleted             timestamp with time zone     null,
    created             timestamp with time zone     not null,
    modified            timestamp with time zone     not null
);

CREATE TABLE roles_teams
(
    id                  uuid primary key           not null,
    role_id             uuid references roles (id) not null,
    team_id             uuid references teams (id) not null,
    created_by_user_id  uuid references users (id) null,
    modified_by_user_id uuid references users (id) null,
    created             timestamp with time zone   not null,
    modified            timestamp with time zone   not null
);

CREATE UNIQUE INDEX roles_teams_pk ON roles_teams (role_id, team_id);

CREATE TABLE roles_users
(
    id                  uuid primary key           not null,
    role_id             uuid references roles (id) not null,
    user_id             uuid references users (id) not null,
    created_by_user_id  uuid references users (id) null,
    modified_by_user_id uuid references users (id) null,
    created             timestamp with time zone   not null,
    modified            timestamp with time zone   not null
);

CREATE UNIQUE INDEX roles_users_pk ON roles_users (role_id, user_id);

CREATE TABLE sync
(
    id                 uuid primary key           not null,
    created_by_user_id uuid references users (id) null,
    created            timestamp with time zone   not null
);

CREATE TABLE audit_log
(
    id        uuid primary key             not null,
    system_id uuid references systems (id) not null,
    sync_id   uuid references sync (id)    not null,
    user_id   uuid references users (id)   null,
    team_id   uuid references teams (id)   null,
    action    varchar                      not null, -- CRUD action
    status    int                          not null, -- Exit code of operation
    message   varchar                      not null, -- User readable success or error message (log line)
    created   timestamp with time zone     not null
);

CREATE INDEX audit_log_created ON audit_log (created);
CREATE INDEX audit_log_action ON audit_log (action);
