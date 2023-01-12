BEGIN;

ALTER TABLE teams RENAME column slack_alerts_channel TO slack_channel;

COMMIT;
