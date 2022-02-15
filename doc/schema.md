# console.nais.io data model

## global for all tables
- created and modified timestamp and user
- soft delete

## users
- id
- e-mail address
- name

pivots:
- teams
- roles
- console roles

## teams
- slug (unique)
- name
- purpose

pivots:
- roles

purpose: nais team provides a platform for all developers in nav

## team k/v data
available via api, to allow data model extension
- id
- slug (foreign key)
- key
- value (text/binary)

## roles (virtual)
- system (console, gcp, etc.)
- resource (maybe?)
- access level
- allow/deny
- child roles?

### systems
- kubernetes
- github
- gcp
- console:create_team

## console roles

- can create team?
- can administer all teams?
- can elevate themselves within teams?
- can delete their own team?
- can delete other teams?
- can assign console roles?
- can assign roles?

## audit log
- foreign key (loosely coupled)
- table name
- human-readable action


| what                | done | date       |
|---------------------|------|------------|
| create namespace    | [x]  | 2022-xxxxx |
| remove user foo@bar | [x]  | 2022-xxxxx |


# Glossary

Partner, customer?
One instance per ^


# Dos and don'ts

- Do act as master
- Do allow manual trigger of team sync
- Do allow partner admins to contact nais team

- Don't assume external state
- Don't sync continuously
