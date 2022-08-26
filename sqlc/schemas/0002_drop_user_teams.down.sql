BEGIN;

CREATE TABLE user_teams (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,

    PRIMARY KEY (user_id, team_id)
);

INSERT INTO user_teams(user_id, team_id)
 (SELECT DISTINCT r.user_id, r.target_id
    FROM user_roles AS r
    JOIN teams AS t ON t.id = r.target_id
    WHERE r.role_name IN ('Team member', 'Team owner'));

COMMIT;