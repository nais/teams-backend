BEGIN;

ALTER TABLE teams ADD COLUMN slack_alerts_channel TEXT;

COMMIT;
