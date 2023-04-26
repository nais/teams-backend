BEGIN;

ALTER TYPE system_name ADD VALUE IF NOT EXISTS 'nais:dependencytrack';

COMMIT;