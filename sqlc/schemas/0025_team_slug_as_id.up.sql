BEGIN;

/* team_metadata */

ALTER TABLE team_metadata
    ADD COLUMN team_slug TEXT;

UPDATE team_metadata
SET team_slug = (
    SELECT slug
    FROM teams
    WHERE teams.id = team_metadata.team_id
);

ALTER TABLE team_metadata
    ALTER COLUMN team_slug SET NOT NULL,
    ADD CONSTRAINT team_metadata_team_slug_fkey FOREIGN KEY(team_slug) REFERENCES teams(slug) ON DELETE CASCADE,
    DROP COLUMN team_id,
    ADD PRIMARY KEY(team_slug, key);

/* reconciler_states */

ALTER TABLE reconciler_states
    ADD COLUMN team_slug TEXT;

UPDATE reconciler_states
SET team_slug = (
    SELECT slug
    FROM teams
    WHERE teams.id = reconciler_states.team_id
);

ALTER TABLE reconciler_states
    ALTER COLUMN team_slug SET NOT NULL,
    ADD CONSTRAINT reconciler_states_team_slug_fkey FOREIGN KEY(team_slug) REFERENCES teams(slug) ON DELETE CASCADE,
    DROP COLUMN team_id,
    ADD PRIMARY KEY(reconciler, team_slug);

/* reconciler_errors */

ALTER TABLE reconciler_errors
    ADD COLUMN team_slug TEXT;

UPDATE reconciler_errors
SET team_slug = (
    SELECT slug
    FROM teams
    WHERE teams.id = reconciler_errors.team_id
);

ALTER TABLE reconciler_errors
    ALTER COLUMN team_slug SET NOT NULL,
    ADD CONSTRAINT reconciler_errors_team_slug_fkey FOREIGN KEY(team_slug) REFERENCES teams(slug) ON DELETE CASCADE,
    DROP COLUMN team_id,
    ADD CONSTRAINT reconciler_errors_team_id_reconciler_key UNIQUE(team_slug, reconciler);

/* user_roles */

ALTER TABLE user_roles
    ADD COLUMN target_team_slug TEXT REFERENCES teams(slug) ON DELETE CASCADE,
    ADD COLUMN target_service_account_id UUID REFERENCES service_accounts(id) ON DELETE CASCADE;

UPDATE user_roles SET target_team_slug = (SELECT slug FROM teams WHERE id = user_roles.target_id) WHERE target_id IS NOT NULL;
UPDATE user_roles SET target_service_account_id = (SELECT id FROM service_accounts WHERE id = user_roles.target_id) WHERE target_id IS NOT NULL;

ALTER TABLE user_roles
    DROP COLUMN target_id;

/* service_account_roles */

ALTER TABLE service_account_roles
    ADD COLUMN target_team_slug TEXT REFERENCES teams(slug) ON DELETE CASCADE,
    ADD COLUMN target_service_account_id UUID REFERENCES service_accounts(id) ON DELETE CASCADE;

UPDATE service_account_roles SET target_team_slug = (SELECT slug FROM teams WHERE id = service_account_roles.target_id) WHERE target_id IS NOT NULL;
UPDATE service_account_roles SET target_service_account_id = (SELECT id FROM service_accounts WHERE id = service_account_roles.target_id) WHERE target_id IS NOT NULL;

ALTER TABLE service_account_roles
    DROP COLUMN target_id;

/* teams */
/*
ALTER TABLE teams
    DROP COLUMN id,
    ADD PRIMARY KEY(slug);


 */


COMMIT;