BEGIN;

ALTER TABLE teams DROP COLUMN slack_alerts_channel;

COMMIT;
