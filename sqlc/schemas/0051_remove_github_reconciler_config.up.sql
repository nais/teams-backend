BEGIN;

DELETE FROM reconciler_config WHERE key IN ('github:org', 'github:app_installation_id');
ALTER TYPE reconciler_config_key RENAME TO reconciler_config_key_old;
CREATE TYPE reconciler_config_key AS ENUM (
    'azure:client_id',
    'azure:client_secret',
    'azure:tenant_id'
);
ALTER TABLE reconciler_config ALTER COLUMN key TYPE reconciler_config_key USING key::text::reconciler_config_key;
DROP TYPE reconciler_config_key_old;

COMMIT;
