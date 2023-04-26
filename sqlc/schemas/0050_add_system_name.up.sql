BEGIN;

ALTER TYPE system_name ADD VALUE IF NOT EXISTS 'nais:deploy';
ALTER TYPE system_name ADD VALUE IF NOT EXISTS 'google:gcp:gar';

COMMIT;