BEGIN;

UPDATE users SET email = LOWER(email);

COMMIT;