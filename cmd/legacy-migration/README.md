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

Synkroniser brukere fra GCP og verifiser at brukere ikke blir fjernet fra databasen.

Ta databasedump, overfør config og data til produksjon, og ta helg.

## Sanere eksisterende løsninger

- navikt/teams
  - Skrive i docs.nais.io om Console vs navikt/teams
  - Erstatte README med notis om å gå til Console
- nais/teams
  - FIXME: aiven networkpolicy, jhrv & co jobber med det nå
  - FIXME: opprettelse av namespace blir ikke gjort enda i legacy-gcp (kanskje lage ny tenant-type, starten av uka)
  - FIXME: securelogs må kanskje utredes om igjen
  - FIXME: docker credentials
  - FIXME: diverse ressurser
  - opprettelse av namespace [krampl: OK]
  - rolebinding med rettigheter samt riktig azure-gruppe [krampl: OK]
  - ca-certificates [kimt: dette er gjort]
  - noe vi har glemt? [vegar: garantert]
  - snorlax utgår
- navikt/google-group-sync
  - Arkiveres, trenger ikke oppdateres
- rbac-sync
  - Se om det gruppe-auth kan skrus på for legacy [krampl]
- ToBAC (den bruker det gamle opplegget med azure ad-"morgruppa")
  - den utgår totalt. Varsle i #nais-announcements om at den skal skrus av asap
  - kanskje den allerede er skrudd av
- "Teams management"-Slack-boten
  - Den er skrevet om til Go + Console. Mangler deploy.
  - Eksisterende workflow ligger i navikt/teams og blir stoppet.
  - Ikke nødvendig for migrering, men bør gjøres snart etterpå.
- alertmanager-config
  - FIXME: Hackes inn i naisd? [krampl]
  - FIXME: legge inn felt i console for slack-alert-channel
  - Oppsett er fortsatt under endring, trenger robusthet og HA
  - mulig for oppretting fra Fasit sin side, med predefinerte navn på alertkanaler: #TEAM-"alerts"-ENV.
  - Kan til slutt vises fram i console-frontend
- azure teams management
  - Slette appen
  - La gruppene være, kommunisere at de ikke er managed lengre