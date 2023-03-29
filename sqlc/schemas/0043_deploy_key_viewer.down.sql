BEGIN;

DELETE FROM role_authz WHERE authz_name='deploy_key:view');

COMMIT;
