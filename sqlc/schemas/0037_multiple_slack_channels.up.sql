BEGIN;

CREATE TABLE slack_alerts_channels (
    team_slug TEXT NOT NULL,
    environment TEXT NOT NULL,
    channel_name TEXT NOT NULL CONSTRAINT slack_alerts_channels_channel_name CHECK (channel_name ~ '^#[a-z0-9æøå_-]{2,80}$'),
    PRIMARY KEY(team_slug, environment)
);

ALTER TABLE teams RENAME column slack_alerts_channel TO slack_channel;

COMMIT;
