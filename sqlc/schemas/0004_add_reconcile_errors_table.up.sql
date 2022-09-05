BEGIN;

CREATE TABLE reconcile_errors (
    id BIGSERIAL PRIMARY KEY,
    correlation_id UUID NOT NULL,
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    system_name system_name NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    error_message TEXT NOT NULL,

    UNIQUE (team_id, system_name)
);

CREATE INDEX idx_reconcile_errors_created_at_desc ON reconcile_errors (created_at DESC);

COMMIT;