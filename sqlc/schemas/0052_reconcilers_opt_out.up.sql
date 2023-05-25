BEGIN;

CREATE TABLE reconciler_opt_outs (
    team_slug text NOT NULL REFERENCES teams(slug) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reconciler_name reconciler_name NOT NULL REFERENCES reconcilers(name) ON DELETE CASCADE,
    PRIMARY KEY(team_slug, user_id, reconciler_name)
);

COMMIT;
