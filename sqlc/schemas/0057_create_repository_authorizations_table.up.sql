BEGIN;

CREATE TYPE repository_authorization_enum AS ENUM (
    'deploy'
);

CREATE TABLE repository_authorizations (
    team_slug text NOT NULL,
    github_repository text NOT NULL,
    repository_authorization repository_authorization_enum NOT NULL,
    PRIMARY KEY(team_slug, github_repository, repository_authorization),
    CHECK ((team_slug ~ '^(?=.{3,30}$)[a-z](-?[a-z0-9]+)+$'::text))
);

ALTER TABLE repository_authorizations
    ADD FOREIGN KEY (team_slug) REFERENCES teams(slug) ON DELETE CASCADE;

COMMIT;
