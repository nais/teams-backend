# ADR - Reconcilere i NAIS Console

**TL;DR** *En reconciler er ansvarlig for å synkronisere et team i NAIS Console ut til et eksternt system (f.eks GitHub eller Azure AD). Console vil ha flere reconcilere, og alle opererer etter "[eventually consistent](https://en.wikipedia.org/wiki/Eventual_consistency)" modellen.*

## Bakgrunn

NAIS Console (**Console**) er APIet som sluttbrukere skal forholde seg til når de skal opprette og vedlikeholde sine team(s), og da gjerne via en [frontend](https://github.com/nais/console-frontend). Et team kan ha koblinger til ett eller flere eksterne systemer, f.eks GitHub, Azure AD eller GCP. Console er ansvarlig for å vedlikeholde disse eksterne ressursene slik at de så langt det lar seg gjøre speiler informasjonen som finnes om teamet i Console.

Dette må implementeres på en oversiktlig og enhetlig måte slik at man enkelt kan introdusere integrasjoner til nye eksterne systemer i fremtiden, samt vedlike en eksisterende integrasjon dersom f.eks et eksternt API endres.

Med tanke på tilgangene som integrasjonene til de eksterne systemene trenger er det også viktig at Console ikke tar over kontrollen på ressurser den ikke i utgangspunktet "eier". Med eierskap så menes eksterne ressurser som Console selv har opprettet.

Vi har valgt å bruke et konsept vi kaller reconcilere, der *en* reconciler er ansvarlig for å vedlikeholde team-koblinger i *ett* eksternt system. Reconcilerne skal i utgangspunktet kunne kjøre i hvilken som helst rekkefølge, og alle følger "eventually consistent"-modellen. Eksempler på reconcilere som allerede er implementert i Console er som følger:

- **GitHub team**: Reconcileren vil opprette et GitHub team i en GitHub organisasjon for Console-teamet, samt speile medlemslisten som ligger i Console så langt det lar seg gjøre. Dersom en bruker i Console ikke finnes i organisasjonen på GitHub er det ikke reconcileren, ei heller Console, sin jobb å opprette brukere. Dette må gjøres av tenanten på andre måter.
- **Azure AD group**: Reconcileren vil opprette en gruppe i en Azure AD tenant for Console-teamet, samt speile medlemslisten som ligger i Console så langt det lar seg gjøre. Manglende brukere i Azure AD blir ikke opprettet av Console eller reconcileren.
- **GCP project**: Reconcileren vil opprette et eller flere prosjekter i GCP for Console teamet.

Alle reconcilere er ansvarlige for å produsere gode logger som kan brukes for å spore endringer over tid (audit logger), samt gi en god forståelse for sluttbruker dersom et Console-team ikke kan synkroniseres til et eksternt system. Noen eksempler på mulige problemer kan være som følger:

- Det ønskede teamnavnet på GitHub er ikke tilgjengelig.
- Det ønskede gruppenavnet i Azure AD er ikke tilgjengelig.
- APIet til GCP er midlertidig utilgjengelig.

## Løsningsbeskrivelse

Console vil ha en kø for team som skal synkroniseres. For hvert team i køen blir det opprettet en unik correlation ID, og denne IDen, i tillegg til teamet, blir sendt til samtlige konfigurerte reconcilere, en etter en, slik at teamet blir synkronisert ut til de eksterne systemene (GitHub, Azure AD osv). Når det første teamet er ferdig synkronisert, vil Console fortsette med neste team på køen. Når køen er tom vil Console vente til det dukker opp et team på køen.

Hver reconciler skal altså ta imot en input som består av to elementer:

- **Et correlation objekt**: Dette objektet refererer til en synkronisering for et team. Objektet inneholder en ID som kan brukes i logger, og andre elementer vi ønsker å knytte sammen med synkroniseringen, f.eks feilmeldinger.
- **Et team objekt**: Det faktiske teamet som skal synkroniseres til de forskjellige eksterne systemene.

Alle reconcilere i den samme synkroniseringen vil motta det samme correlation og team objektet, og hver reconciler ender opp med å returnere en feil dersom noe gjør at den ikke kan fullføre. Dersom den ikke møter på noen problemer underveis vil den ikke returnere noe. Et team blir bare tatt av synkroniseringskøen dersom alle reconcilere fullfører uten feil for det spesififkke teamet. Dersom minst en reconciler feiler forblir teamet på køen, og en ny runde blir startet etter en kort pause. Det skal med andre ord ikke være nødvendig å trigge en sync manuelt dersom noe gikk galt i forrige synkronisering.

### Konfigurasjon av reconcilere

En tenant kan selv bestemme hvilke reconcilere som skal benyttes. Alle kan aktiveres via miljøvariabler, og alle kan ha forskjellige konfigurasjonsparametre. Azure AD reconcileren trenger nøkler for å kunne prate med Microsoft Graph API, mens GitHub reconcileren trenger en rekke andre parametre. Alt dette må dokumenteres i Console slik at tenanten vet hva som trengs for de forskjellige reconcilerne.

### Synkronisering ved oppstart

Når Console starter opp vil den starte en komplett synkronisering av samtlige team i databasen. Dette skjer ved hver oppstart av Console. 

### Kontinuerlig synkronisering

En synkronisering av et team startes etter følgende hendelser i Console:

- Når teamet blir opprettet
- Når teamet endres (f.eks medlemmer blir fjernet / lagt til)
- Når teamet slettes
- Sluttbruker ber om en manuell synkronisering via APIet. Dette kan være nødvendig dersom en bruker f.eks har slettet den eksterne ressursen ved en feil, og ønsker å få det opprettet igjen.

### Initiell opprettelse av eksterne ressurser

En reconciler kan holde styr på "state" i de eksterne systemene ved behov, og dette lagres i databasen til Console.

Et eksempel på et reconciler som benytter seg av state er GitHub teams. Når Console teamet først blir opprettet kommer GitHub team reconcileren til å gjøre sin første synkronisering av teamet. Den kommer da ikke til å ha noen state for teamet, så den vil forsøke å lage et nytt team i organisasjonen til tenanten på GitHub. Dersom teamnavnet som Console ønsker å opprette ikke er tilgjengelig vil ikke reconcileren klare å gjennomføre synkroniseringen, og en feil blir lagret i databasen knyttet til correlation objektet, som da kan presenteres for sluttbruker slik at vedkommende kan se hvorfor GitHub teamet ikke ble opprettet. Dersom dette skjer blir ingen state for teamet lagret i Console.

Dersom teamnavnet er ledig vil Console få opprettet den eksterne ressursen, og den eksterne IDen til GitHub teamet blir lagret i Console databasen som state. GitHub reconcileren kan da benytte seg av denne IDen ved neste synkronisering av teamet.

### Vedlikehold av eksterne ressurser

Dersom et Console team f.eks endrer medlemmer, vil en ny synkronisering av teamet starte. Om vi bruker GitHub team reconcileren som eksempel igjen vil den da gjøre følgende:

1. Hente den eksterne IDen til teamet fra state, som Console selv tidligere har satt.
2. Forsøke å hente teamet fra GitHub APIet for å se om det fortsatt eksisterer.
   1. Om det av en eller annen grunn ikke lenger eksisterer på GitHub vil Console forsøke å opprette et nytt GitHub team. Dersom dette ikke går avbrytes synkroniseringen for teamet, og Console vil forsøke på nytt etter en kort periode.
   2. Dersom GitHub teamet eksisterer kan vi fortsette synkroniseringen.
3. Legg til / fjern medlemmer i teamet basert på medlemslisten i Console.

Ved å benytte seg av state vil en reconciler da aldri forsøke å ta over et eksternt team som reconcileren selv ikke opprettet. Det eneste unntaket er når state manuelt endres i databasen til Console. 

### Manuell endring av state

Ved spesielle omstendigheter kan det være at vi ønsker å manuelt endre state som ligger lagret i Console. Dette kan være at en tenant allerede har et eller flere team på GitHub, eller en gruppe i Azure AD, som man da ønsker å knytte til det mye teamet i Console. Om dette er tilfelle kan da en bruker med de nødvendige rettighetene gjøre en manuell kobling i Console, slik at reconcileren tror den jobber med en ekstern ressurs den selv har opprettet tidligere. Rutiner for bruk av denne metoden er det tenanten selv som må definere.

### Sletting av eksterne ressurser

Når en bruker sletter et team i Console vil ikke alle tilknyttede eksterne ressurser nødvendigvis slettes. Sluttbrukere må få et valg om de ønsker å slette eksterne ressurser eller ei. Etter en bruker har tatt et aktiv valg vil en reconciler så langt det lar seg gjøre slette eksterne ressurser.