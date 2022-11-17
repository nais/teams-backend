BEGIN;

DELETE FROM reconcilers WHERE name = 'nais:deploy-key';

COMMIT;