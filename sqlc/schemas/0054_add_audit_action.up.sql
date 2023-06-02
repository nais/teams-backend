BEGIN;

ALTER TYPE audit_action RENAME VALUE 'dependencytrack:group:create' TO 'dependencytrack:team:create';
ALTER TYPE audit_action ADD VALUE 'dependencytrack:team:add-member' BEFORE 'dependencytrack:team:create';
ALTER TYPE audit_action ADD VALUE 'dependencytrack:team:delete-member' AFTER 'dependencytrack:team:create';

COMMIT;