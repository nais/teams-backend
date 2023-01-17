BEGIN;

DROP TABLE slack_alerts_channels;
ALTER TABLE teams RENAME column slack_channel TO slack_alerts_channel;

COMMIT;