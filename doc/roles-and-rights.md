Roles and rights
================

Actions
-------

- create team
- delete team
- 
- add user to team
- remove user from team
- 
- assign role to user
- remove role from user
- 
- view team list
- view any team
- view own team
- view any user
- view own user
- 
- create service account
- read service account
- update service account
- delete service account
- 
- view complete auditlog 
- 
- create own api-key
- create api-key for specific user
- create any api-key
- delete own api-key
- delete api-key for specific user
- delete any api-key
- 
- Add team property (key/value)
- Remove team property
- Edit team property
- 
- Add team system state
- Remove team system state
- Edit team system state

Proposal 1
----------

### Role

```yaml
rules:
  - target:
      kind: "team", "team/user", "team/property", "user", "user/apikey", "rolebinding"
      name: Name, ID or wildcard
    action: "get", "list", "update", "delete", "create"
  ...
```

### Role Binding

```yaml
role: Reference to role
subjects:
  - name: username, userID or wildcard
  ...
```

### Notes

- Who makes the Roles?
- (Too?) flexible
- Target specification too broad or too limited
