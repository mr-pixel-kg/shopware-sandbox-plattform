# API Backend

Pragmatisches Backend fuer die Docker-Sandbox-Plattform mit Echo, GORM, PostgreSQL, JWT und Gast-Session-Cookie.

## Konfiguration

- `config.example.yml` nach `config.yml` kopieren
- Optional `CONFIG_PATH=/pfad/zur/config.yml` setzen

## Start

1. PostgreSQL starten
2. Migration `internal/database/migrations/000001_init.sql` ausfuehren
3. API starten:

```bash
go run ./cmd/api
```

## Wichtige Punkte

- Oeffentliche Gaeste erhalten automatisch ein Cookie und koennen darueber ihre eigenen Sandboxes wiedersehen.
- Mitarbeiter erhalten JWTs und koennen eigene sowie alle aktiven Sandboxes sehen.
- Limits greifen global, pro IP fuer Gaeste und pro User fuer eingeloggte Mitarbeiter.
- Swagger/OpenAPI ist unter `/swagger` und `/swagger/openapi.yaml` verfuegbar.
- Fuer Traefik sollten `docker.network`, `docker.traefik_*` und `sandbox.internal_port` passend zur Laufzeitumgebung gesetzt werden.
