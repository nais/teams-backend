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
    go run cmd/legacy-migration/main.go 2>&1 | tee local/migration.log
grep -E "(WARN|ERR)" local/migration.log | grep -v "srvvsb@navno"
```

(obsolete, nye teams settes disabled av synkroniseringsscriptet)
Sett alle teams som disabled (skal ikke synkes):

```
PGPASSWORD=console psql -h localhost -p 3002 -U console console -c 'update teams set enabled = false;'
```

## Klargjøre for migrering

Logg inn med `nais.io`-bruker:

```
gcloud auth login --update-adc
```

Hent console OAuth2-credentials fra [Fasit](https://fasit.nais.io/tenant/nav/management?feature=console&tab=helm_values).

Legg inn i `local/production.env`:
```
CONSOLE_GOOGLE_MANAGEMENT_PROJECT_ID=nais-management-233d
CONSOLE_OAUTH_CLIENT_ID=xxx
CONSOLE_OAUTH_CLIENT_SECRET=xxx
CONSOLE_OAUTH_REDIRECT_URL=http://localhost:3000/oauth2/callback
CONSOLE_TENANT_DOMAIN=nav.no
CONSOLE_USERSYNC_ENABLED=true
```

Synkroniser brukere fra GCP og verifiser at brukere ikke blir fjernet fra databasen.

## Verifikasjon

Sjekk at det ikke eksisterer tomme teams:

```sql
select t.slug, count(ur.id)
from teams as t
    left join user_roles as ur
        on (t.slug = ur.target_team_slug)
group by t.slug
order by count(ur.id);
```

Sjekk at det er tildelt cirka-ish antall roller:

```sql
select count(*) from users;
select count(role_name), role_name from user_roles group by role_name;
```

Sjekk at teamene finnes:

```sql
select slug,slack_alerts_channel from teams order by slug asc;
```

Forsikre deg om at alle reconcilers er skrudd av:

```sql
select * from reconcilers where enabled = true;
-- name | display_name | description | enabled | run_order
--------+--------------+-------------+---------+-----------
-- (0 rows)
```

## Overføring til produksjon

Ta databasedump, overfør config og data til produksjon, og ta helg.

```
PGPASSWORD=console pg_dump --clean --if-exists --no-owner -h localhost -p 3002 -U console console > local/console-production-import.sql
```

## Sanere eksisterende løsninger

- navikt/teams
  - Skrive i docs.nais.io om Console vs navikt/teams [kimt: laget branch klar til merge]
  - Erstatte README med notis om å gå til Console [kimt: laget branch klar til merge]
- nais/teams
  - config connector [OK]
  - networkpolicies [jhrv: OK]
  - opprettelse av namespace blir ikke gjort enda i legacy-gcp [krampl: gjort, men utestet]
  - rolebindings til nais deploy/teams [OK]
  - securelogs fluentd-config [terjes/jhrv: ferdig]
  - docker credentials [terjes/jhrv: ferdig]
  - resourcequota [videreføres ikke med mindre det oppstår behov]
  - opprettelse av namespace [krampl: OK]
  - rolebinding med rettigheter samt riktig azure-gruppe [krampl: OK]
  - ca-certificates [kimt: dette er gjort]
  - noe vi har glemt? [vegar: garantert]
  - snorlax utgår
- alertmanager-config
  - FIXME: legge inn redigeringsmulighet i console for slack-alert-channel -> sendes til naisd
  - Oppsett er fortsatt under endring, trenger robusthet og HA
  - mulig for oppretting fra Fasit sin side, med predefinerte navn på alertkanaler: #TEAM-"alerts"-ENV.
  - Kan til slutt vises fram i console-frontend
- navikt/google-group-sync
  - Arkiveres, trenger ikke oppdateres
- rbac-sync
  - erstattes av gke-security-groups, kan disables etter sync/migrering
  - kan tre inn i on-prem-clustre etterpå [frodes]
- ToBAC (den bruker det gamle opplegget med azure ad-"morgruppa")
  - den utgår totalt. Varsle i #nais-announcements om at den skal skrus av asap
  - kanskje den allerede er skrudd av
- "Teams management"-Slack-boten
  - Den er skrevet om til Go + Console. Mangler deploy.
  - Eksisterende workflow ligger i navikt/teams og blir stoppet.
  - Ikke nødvendig for migrering, men bør gjøres snart etterpå.
- azure teams management
  - Slette appen
  - La gruppene være, kommunisere at de ikke er managed lengre
