BEGIN;

/* PostgreSQL does not support removing single items from an ENUM, one must DROP the type, and re-create it */

DELETE FROM role_authz WHERE
  (role_name='Team owner' AND authz_name='teams:synchronize') OR
  (role_name='Synchronizer' AND authz_name='teams:synchronize') OR
  (role_name='Synchronizer' AND authz_name='usersync:synchronize') OR
  (role_name='Admin' AND authz_name='teams:synchronize') OR
  (role_name='Admin' AND authz_name='usersync:synchronize')
    ;

COMMIT;
