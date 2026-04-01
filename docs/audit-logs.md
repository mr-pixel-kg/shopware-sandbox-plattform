# Audit Logs

Diese Datei beschreibt, wie Audit Logs im Projekt funktionieren und wie neue Audit-Eintraege sauber ergänzt werden.

## Ziel

Audit Logs sollen nachvollziehbar machen:

- wer eine Aktion ausgelöst hat
- auf welche Ressource sich die Aktion bezog
- aus welchem Client-Kontext die Aktion kam
- wann die Aktion passiert ist
- welche relevanten Zusatzdaten es dazu gab

Die Audit Logs sind bewusst nicht als fachliche Event-Sourcing-Historie gedacht. Sie dienen in erster Linie der Nachvollziehbarkeit, Administration und späteren Analyse.

## Datenmodell

Das Datenmodell liegt in [api/internal/models/audit_log.go](/Users/manuel.kienlein/GolandProjects/shopware-testenv-platform/api/internal/models/audit_log.go).

Aktuelle Felder:

- `id`
- `userId`
- `action`
- `ipAddress`
- `userAgent`
- `clientToken`
- `resourceType`
- `resourceId`
- `details`
- `createdAt`

Wichtige Eigenschaften:

- `userId` darf `nil` sein
  Das ist wichtig fuer anonyme oder systemnahe Aktionen.
- `ipAddress` darf `nil` sein
  Nicht jede Anfrage liefert zuverlaessig eine Client-IP.
- `userAgent` darf `nil` sein
  Manche Clients senden keinen User-Agent.
- `clientToken` darf `nil` sein
  Die API ist aktuell nur darauf vorbereitet; die finale Client-Identitaet ist noch nicht fest definiert.
- `resourceType` und `resourceId` duerfen `nil` sein
  Das ist vor allem fuer globale Auth-Aktionen wie Login/Logout sinnvoll.

Die Migration fuer die neuen Audit-Felder liegt in [api/internal/database/migrations/000010_audit_log_actor_and_resource.sql](/Users/manuel.kienlein/GolandProjects/shopware-testenv-platform/api/internal/database/migrations/000010_audit_log_actor_and_resource.sql).

## Zentrale Contracts

Actions und Resource-Typen werden zentral in [api/internal/auditlog/contracts.go](/Users/manuel.kienlein/GolandProjects/shopware-testenv-platform/api/internal/auditlog/contracts.go) definiert.

### Warum zentral?

Vor dem Refactoring wurden Actions als freie Strings an vielen Stellen erzeugt. Das fuehrte leicht zu:

- inkonsistenten Namen
- vermischten Ebenen wie `admin.user_created` vs. `user.created`
- erschwerter Filterbarkeit im UI
- hohem Risiko fuer Tippfehler

Jetzt gilt:

- neue Actions werden zuerst in `contracts.go` definiert
- neue Resource-Typen werden ebenfalls dort definiert
- Call-Sites verwenden diese Konstanten statt freie Strings

## Struktur eines Audit-Eintrags

Der Service verwendet [api/internal/services/audit_service.go](/Users/manuel.kienlein/GolandProjects/shopware-testenv-platform/api/internal/services/audit_service.go).

Zentrales Input-Modell:

```go
type AuditLogInput struct {
    Actor        AuditActor
    Action       auditcontracts.Action
    ResourceType *auditcontracts.ResourceType
    ResourceID   *uuid.UUID
    Details      map[string]any
}
```

Der Actor-Kontext:

```go
type AuditActor struct {
    UserID      *uuid.UUID
    IPAddress   *string
    UserAgent   *string
    ClientToken *uuid.UUID
}
```

## Herkunft des Actor-Kontexts

HTTP-Handler bauen den Audit-Kontext ueber [api/internal/http/handlers/audit_helpers.go](/Users/manuel.kienlein/GolandProjects/shopware-testenv-platform/api/internal/http/handlers/audit_helpers.go).

Dort werden aktuell gesammelt:

- `userId`
- `c.RealIP()`
- `c.Request().UserAgent()`
- optionaler `clientToken`

Der `clientToken` wird aktuell nur passiv gelesen, nicht erzeugt oder gesetzt. Die Middleware dafuer liegt in [api/internal/http/middleware/client_token.go](/Users/manuel.kienlein/GolandProjects/shopware-testenv-platform/api/internal/http/middleware/client_token.go).

Aktueller Stand:

- Header `X-Client-Token` wird gelesen
- alternativ Cookie `client_token` wird gelesen
- die API setzt selbst keinen Cookie
- die finale Produktentscheidung, ob Header oder Cookie verwendet wird, ist noch offen

## Best Practices fuer neue Audit-Eintraege

### 1. ResourceType und ResourceId setzen, wenn eine Ressource betroffen ist

Wenn eine Aktion eine konkrete Ressource erstellt, aendert oder loescht, sollen immer beide Felder gesetzt werden:

- `resourceType`
- `resourceId`

Beispiele:

- Sandbox erstellt: `resourceType=sandbox`, `resourceId=<sandbox uuid>`
- Image geloescht: `resourceType=image`, `resourceId=<image uuid>`
- Benutzer aktualisiert: `resourceType=user`, `resourceId=<user uuid>`

### 2. Actions fachlich benennen, nicht nach Rollen oder UI

Die Action beschreibt die fachliche Aenderung, nicht die Rolle oder den Screen.

Gut:

- `user.created`
- `user.updated`
- `sandbox.deleted`
- `image.thumbnail_uploaded`

Weniger gut:

- `admin.user_created`
- `images.edit_modal_saved`
- `sandbox_button_clicked`

Die Rolle des Ausloesers ist indirekt ueber `userId` und die eigentliche Fachlogik ableitbar. Sie gehoert nicht in den Action-Namen.

### 3. Details nur fuer Zusatzkontext verwenden

`details` soll keine Duplikation von Kernfeldern sein, die schon in:

- `action`
- `resourceType`
- `resourceId`
- `userId`

stecken.

Gut in `details`:

- geaenderte Felder
- neue TTL in Minuten
- Ziel-Image-Name bei Snapshot
- Meta-Informationen zur Art der Aktion

Weniger gut:

- nochmal `resourceId`, wenn sie schon im Hauptfeld steht
- komplette Request-Bodies
- grosse oder sensible Rohdaten

### 4. Audit Logs duerfen nie die Fachlogik blockieren

Aktuell werden Audit-Aufrufe an vielen Stellen bewusst mit `_ = ...` ignoriert. Das ist hier gewollt:

- ein fehlgeschlagener Audit-Eintrag soll z. B. keine Sandbox-Erstellung verhindern
- die Geschaeftsaktion bleibt wichtiger als das Logging

Wenn das spaeter geaendert werden soll, sollte das bewusst und separat entschieden werden.

### 5. Keine sensitiven Daten in Details schreiben

Nicht loggen:

- Passwoerter
- Tokens
- Geheimnisse
- komplette Authorization-Header

Auch bei `clientToken` gilt: Das Feld selbst ist okay, aber keine zusaetzlichen sensitiven Client-Geheimnisse in `details` ablegen.

## Aktuelle Resource-Typen

Aktuell definiert in `contracts.go`:

- `image`
- `sandbox`
- `user`

Wenn weitere fachliche Ressourcen dazukommen, dort erweitern.

## Aktuelle Actions

Aktuell zentral definiert:

### Auth

- `auth.logged_in`
- `auth.logged_out`

### User

- `user.registered`
- `user.created`
- `user.updated`
- `user.deleted`
- `user.whitelisted`
- `user.whitelist_removed`

### Image

- `image.created`
- `image.updated`
- `image.deleted`
- `image.thumbnail_uploaded`
- `image.thumbnail_deleted`
- `image.snapshot_created`

### Sandbox

- `sandbox.created`
- `sandbox.updated`
- `sandbox.ttl_updated`
- `sandbox.deleted`

## Wo aktuell Audit Logs geschrieben werden

Wichtige Call-Sites:

- [api/internal/http/handlers/auth_handler.go](/Users/manuel.kienlein/GolandProjects/shopware-testenv-platform/api/internal/http/handlers/auth_handler.go)
- [api/internal/http/handlers/image_handler.go](/Users/manuel.kienlein/GolandProjects/shopware-testenv-platform/api/internal/http/handlers/image_handler.go)
- [api/internal/http/handlers/user_handler.go](/Users/manuel.kienlein/GolandProjects/shopware-testenv-platform/api/internal/http/handlers/user_handler.go)
- [api/internal/http/handlers/whitelist_handler.go](/Users/manuel.kienlein/GolandProjects/shopware-testenv-platform/api/internal/http/handlers/whitelist_handler.go)
- [api/internal/services/sandbox_service.go](/Users/manuel.kienlein/GolandProjects/shopware-testenv-platform/api/internal/services/sandbox_service.go)

Faustregel:

- Handler loggen eher request-nahe Aktionen
- Services loggen eher fachliche Lifecycle-Aenderungen

Wenn eine Aktion tief in der Fachlogik entsteht, sollte sie eher im Service geloggt werden.

## Wie man eine neue Audit-Action hinzufuegt

1. Neue Action in [api/internal/auditlog/contracts.go](/Users/manuel.kienlein/GolandProjects/shopware-testenv-platform/api/internal/auditlog/contracts.go) definieren.
2. Falls noetig neuen `ResourceType` ebenfalls dort anlegen.
3. An der passenden Call-Site `newAuditLogInput(...)` oder `AuditLogInput{...}` verwenden.
4. `resourceType` und `resourceId` setzen, wenn eine konkrete Ressource betroffen ist.
5. Nur sinnvolle, kleine Zusatzinfos in `details` schreiben.
6. Falls die Action im Admin-UI speziell dargestellt werden soll, Mapping in [web/src/views/admin/AdminAuditView.vue](/Users/manuel.kienlein/GolandProjects/shopware-testenv-platform/web/src/views/admin/AdminAuditView.vue) erweitern.

## API und Frontend

Die Audit-API liefert Audit-Eintraege ueber:

- [api/internal/http/handlers/audit_handler.go](/Users/manuel.kienlein/GolandProjects/shopware-testenv-platform/api/internal/http/handlers/audit_handler.go)

Response-Typ:

- [api/internal/http/dto/responses.go](/Users/manuel.kienlein/GolandProjects/shopware-testenv-platform/api/internal/http/dto/responses.go)

Frontend-Typen und Anzeige:

- [web/src/types/api.types.ts](/Users/manuel.kienlein/GolandProjects/shopware-testenv-platform/web/src/types/api.types.ts)
- [web/src/composables/useAuditLogs.ts](/Users/manuel.kienlein/GolandProjects/shopware-testenv-platform/web/src/composables/useAuditLogs.ts)
- [web/src/views/admin/AdminAuditView.vue](/Users/manuel.kienlein/GolandProjects/shopware-testenv-platform/web/src/views/admin/AdminAuditView.vue)

## Audit-Log-API

Aktueller Endpoint:

- `GET /api/audit-logs`

Die Response bleibt bewusst ein einfaches Array von Audit-Log-Eintraegen, damit bestehende Clients nicht brechen.

### Query-Parameter

- `limit`
  Standard `50`, maximal `500`
- `offset`
  Standard `0`
- `userId`
  Filter auf einen konkreten Benutzer
- `action`
  Filter auf eine konkrete Action
- `resourceType`
  Filter auf einen konkreten Ressourcentyp
- `resourceId`
  Filter auf eine konkrete Ressource
- `clientToken`
  Filter auf einen konkreten Client-Token
- `from`
  RFC3339-Zeitstempel, inklusiver Startzeitpunkt
- `to`
  RFC3339-Zeitstempel, inklusiver Endzeitpunkt

Beispiel:

```text
/api/audit-logs?resourceType=sandbox&action=sandbox.deleted&from=2026-04-01T00:00:00Z&limit=100
```

### Validierung

Die API liefert `400 VALIDATION_ERROR`, wenn:

- `limit` ausserhalb des erlaubten Bereichs liegt
- `offset` negativ ist
- `userId`, `resourceId` oder `clientToken` keine UUID ist
- `from` oder `to` kein gueltiger RFC3339-Zeitstempel ist
- `from` spaeter als `to` ist

## Offene Punkte

### Client Token

Der `clientToken` ist im Datenmodell vorgesehen, aber produktseitig noch nicht final entschieden.

Offen ist:

- kommt er per Header oder Cookie?
- wer erzeugt ihn?
- wie stabil und vertrauenswuerdig ist er?
- soll die API ihn spaeter selbst setzen oder nur konsumieren?

Bis das entschieden ist, bleibt die Middleware bewusst passiv.

### Filter und Suche

Das Datenmodell unterstuetzt jetzt gute spaetere Filter nach:

- `action`
- `resourceType`
- `resourceId`
- `clientToken`
- `userId`
- Zeitraum

Die List-API nutzt diese Moeglichkeiten aktuell noch nicht voll aus.
