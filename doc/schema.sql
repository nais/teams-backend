-- TODO:
-- - child roles
-- - pivot tables

CREATE TABLE users
(
    id                  varchar primary key           not null,
    email               varchar unique                null,
    apikey              varchar unique                null,
    name                varchar                       not null,
    deleted_by_user_id  varchar references users (id) null,
    created_by_user_id  varchar references users (id) null,
    modified_by_user_id varchar references users (id) null,
    deleted             timestamp with time zone      null,
    created             timestamp with time zone      not null,
    modified            timestamp with time zone      not null
);

CREATE TABLE teams
(
    id                  varchar primary key           not null,
    name                varchar unique                not null,
    purpose             varchar                       null,
    deleted_by_user_id  varchar references users (id) null,
    created_by_user_id  varchar references users (id) null,
    modified_by_user_id varchar references users (id) null,
    deleted             timestamp with time zone      null,
    created             timestamp with time zone      not null,
    modified            timestamp with time zone      not null
);

CREATE TABLE team_data
(
    id                  varchar primary key           not null,
    team_id             varchar                       not null,
    key                 varchar                       not null,
    value               varchar                       null,
    deleted_by_user_id  varchar references users (id) null,
    created_by_user_id  varchar references users (id) null,
    modified_by_user_id varchar references users (id) null,
    deleted             timestamp with time zone      null,
    created             timestamp with time zone      not null,
    modified            timestamp with time zone      not null
);

CREATE UNIQUE INDEX team_data_pk ON team_data (team_id, key);

CREATE TABLE systems
(
    id                  varchar primary key           not null,
    deleted_by_user_id  varchar references users (id) null,
    created_by_user_id  varchar references users (id) null,
    modified_by_user_id varchar references users (id) null,
    deleted             timestamp with time zone      null,
    created             timestamp with time zone      not null,
    modified            timestamp with time zone      not null
);

CREATE TABLE roles
(
    id                  varchar primary key             not null,
    system_id           varchar references systems (id) not null,
    resource            varchar                         not null, -- maybe
    access_level        varchar                         not null,
    permission          varchar                         not null, -- allow/deny
    deleted_by_user_id  varchar references users (id)   null,
    created_by_user_id  varchar references users (id)   null,
    modified_by_user_id varchar references users (id)   null,
    deleted             timestamp with time zone        null,
    created             timestamp with time zone        not null,
    modified            timestamp with time zone        not null
);

CREATE TABLE sync
(
    id                 varchar primary key           not null,
    created_by_user_id varchar references users (id) null,
    created            timestamp with time zone      not null
);

CREATE TABLE audit_log
(
    id          varchar primary key             not null,
    system_id   varchar references systems (id) not null,
    action      varchar                         not null,
    sync_id     varchar references sync (id)    not null,
    user_id     varchar references users (id)   null,
    team_id     varchar references teams (id)   null,
    description varchar                         not null,
    created     timestamp with time zone        not null
);