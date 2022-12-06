BEGIN;

ALTER TABLE teams ADD COLUMN slack_alerts_channel TEXT NOT NULL CONSTRAINT teams_slack_alerts_channel CHECK (slack_alerts_channel ~ '^#[a-z0-9æøå_-]{2,80}$');

COMMIT;
