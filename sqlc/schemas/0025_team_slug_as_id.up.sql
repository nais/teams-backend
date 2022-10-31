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

COMMIT;