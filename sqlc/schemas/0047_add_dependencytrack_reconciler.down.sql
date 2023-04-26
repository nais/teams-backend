/* PostgreSQL does not support removing single items from an ENUM, one must DROP the type, and re-create it */
BEGIN;

DELETE FROM reconcilers WHERE name = 'nais:dependencytrack';

COMMIT;
