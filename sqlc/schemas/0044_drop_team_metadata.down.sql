BEGIN;

CREATE TABLE team_metadata (
    team_slug TEXT;
    key TEXT NOT NULL,
    value TEXT,

    PRIMARY KEY (team_id, key)
);

ALTER TABLE team_metadata
    ALTER COLUMN team_slug SET NOT NULL,
    ADD CONSTRAINT team_metadata_team_slug_fkey FOREIGN KEY(team_slug) REFERENCES teams(slug) ON DELETE CASCADE ;

COMMIT;
