BEGIN;

DELETE FROM reconciler_states
WHERE reconciler = 'nais:dependencytrack';

COMMIT;
