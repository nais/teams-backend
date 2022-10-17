# Playbook for migrering

## Sett opp datagrunnlag

```
# Clone og pull nais/teams og navikt/teams.
...

# Kopier inn datagrunnlag
% cp ~/src/navikt/teams/teams.{yml,json} local/
% cp ~/src/nais/teams/gcp-projects/{dev,prod}-output.json local/gcp-cache/

# Slett databasen og opprett den igjen
% docker-compose down -v
% docker-compose up
```

Legg følgende inn i `local/migration.env`

```
CONSOLE_IMPORTER_AZURE_CLIENT_ID=xxx
CONSOLE_IMPORTER_AZURE_CLIENT_SECRET=xxx
CONSOLE_IMPORTER_AZURE_TENANT_ID=62366534-1ec3-4962-8869-9b5535279d0b
CONSOLE_TENANT_DOMAIN=nav.no
```

Importer real-world data inn i databasen og sjekk for feil:

```
env $(cat local/migration.env | xargs) \
    go run cmd/legacy-migration/main.go | tee local/migration.log
grep -E "(WARN|ERR)" local/migration.log | grep -v "srvvsb@navno"
```

Sett alle teams som disabled (skal ikke synkes):

```
PGPASSWORD=console psql -h localhost -p 3002 -U console console -c 'update teams set enabled = false;'
```

## Klargjøre for migrering

Hent console GCP-credentials fra [Fasit](https://fasit.nais.io/tenant/nav/management?feature=console&tab=helm_values)
og base64-decode inn i `local/credentials.json`.

Legg inn i `local/production.env`:
```
CONSOLE_TENANT_DOMAIN=nav.no
CONSOLE_GOOGLE_CREDENTIALS_FILE=local/credentials.json
CONSOLE_GOOGLE_DELEGATED_USER=nais-console@nav.no
CONSOLE_NAIS_NAMESPACE_PROJECT_ID=nais-management-7178
CONSOLE_OAUTH_CLIENT_ID=xxx
CONSOLE_OAUTH_CLIENT_SECRET=xxx
CONSOLE_OAUTH_REDIRECT_URL=http://localhost:3000/oauth2/callback
```

FIXME: legge delegated user-enven inn i Fasit

Synkroniser brukere fra GCP og verifiser at brukere ikke blir fjernet fra databasen.

Arkiver/disable eksisterende løsninger:
- navikt/teams
- nais/teams
  - opprettelse av namespace [jhrv: antakelig OK]
  - rolebinding med rettigheter samt riktig azure-gruppe [jhrv: MÅ vi bruke azure on-prem?]
  - ca-certificates [jhrv: hva er fremtiden?] [kimt: i hvilken form skal dette videreføres for NAV?]
  - noe vi har glemt?
- alertmanager-config opprettes av Fasit, med predefinerte navn på alertkanaler: #TEAM-"alerts"-ENV. Vises fram i console-frontend. [jhrv: OK]
- navikt/google-group-sync
- rbac-sync
- ToBAC (den bruker det gamle opplegget med azure ad-"morgruppa")
  - den utgår totalt. Varsle i #nais-announcements om at den skal skrus av asap
- "Teams management"-Slack-boten