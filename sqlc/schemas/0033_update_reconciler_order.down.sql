BEGIN;

UPDATE reconcilers SET run_order = 101 WHERE run_order = 1;
UPDATE reconcilers SET run_order = 102 WHERE run_order = 2;
UPDATE reconcilers SET run_order = 103 WHERE run_order = 3;
UPDATE reconcilers SET run_order = 104 WHERE run_order = 4;
UPDATE reconcilers SET run_order = 105 WHERE run_order = 5;
UPDATE reconcilers SET run_order = 106 WHERE run_order = 6;

UPDATE reconcilers SET run_order = 1 WHERE name = 'google:workspace-admin';
UPDATE reconcilers SET run_order = 2 WHERE name = 'google:gcp:project';
UPDATE reconcilers SET run_order = 3 WHERE name = 'nais:namespace';
UPDATE reconcilers SET run_order = 4 WHERE name = 'azure:group';
UPDATE reconcilers SET run_order = 5 WHERE name = 'github:team';
UPDATE reconcilers SET run_order = 6 WHERE name = 'nais:deploy';

COMMIT;
