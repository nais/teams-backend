BEGIN;

/* PostgreSQL does not support removing single items from an ENUM, one must DROP the type, and re-create it */

DELETE FROM role_authz WHERE
  (role_name='NaisTeam creator' AND authz_name='teams:skip_nais_validation') OR
  (role_name='NaisTeam creator' AND authz_name='teams:create')
    ;

COMMIT;
