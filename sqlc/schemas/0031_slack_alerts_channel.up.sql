BEGIN;

ALTER TABLE teams ADD COLUMN slack_alerts_channel TEXT CONSTRAINT teams_slack_alerts_channel CHECK (slack_alerts_channel ~ '^#[a-z0-9æøå_-]{2,80}$');
UPDATE teams SET slack_alerts_channel = '#' || slug;
ALTER TABLE teams ALTER COLUMN slack_alerts_channel SET NOT NULL;

COMMIT;
