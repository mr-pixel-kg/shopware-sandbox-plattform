# mpXsandbox

This application allows you to create demo shops in a docker environment.

## Quick Start

### 1. Setup configuration

```bash
cp .env.example .env                      # edit docker socket path and db credentials
cp api/config.example.yml api/config.yml  # edit with your API settings
```

Or using [bab](https://bab.sh):

```bash
bab setup
```

### 2. Start development

```bash
bab dev
```

This starts the database via Docker, then runs the API and web frontend dev servers in parallel.

### 3. Access

- **Web Frontend:** http://localhost:5173
- **API:** http://localhost:8080

## Architecture

| Component | Directory | Stack                                |
|-----------|-----------|--------------------------------------|
| Api       | `/api/`   | Go, Echo, GORM, PostgreSQL, JWT Auth |
| Web       | `/web/`   | Vue 3, Shadcn-vue, TypeScript, pnpm  |

## Configuration Guide

### Configuration File (`api/config.yml`)

Copy `api/config.example.yml` to `api/config.yml` and edit with your settings.
The example is pre-configured for local development (localhost database, dev JWT secret, Traefik disabled).

```yaml
logging:
  level: "debug"
  format: "text"

server:
  port: 8080
  app_url: "http://localhost:8080"
  allowed_origins:
    - "http://localhost:5173"
    - "http://localhost:8000"

database:
  host: "localhost"
  port: 5432
  user: "mrpix_sandbox"
  password: "sandXbox_mrpix2025"
  name: "mrpix_sandbox"
  sslmode: "disable"

auth:
  jwt_secret: "local-dev-secret"
  jwt_ttl_minutes: 480
  guest_jwt_ttl_minutes: 43200
  guest_cookie_name: "shopshredder_guest"

sandbox:
  url_prefix: "sandbox-"
  url_suffix: ".localhost"
  default_lifetime: 60
  max_lifetime: 1440
  cleanup_interval_seconds: 60
  internal_port: 80

docker:
  mode: "port"
  network: "internal"
  traefik_enable: false
  traefik_entrypoints: "websecure"
  traefik_certresolver: "production"
  traefik_middlewares: "sandbox-middleware@file,https-redirect@file"
  snapshot_author: "shopshredder-api"
  snapshot_comment: "Sandbox snapshot created by Shopshredder API"

storage:
  thumbnail_dir: "storage/thumbnails"

guard:
  max_total_sandboxes: 32
  max_sandboxes_per_ip: 5
  max_sandboxes_per_user: 10
```

### Configuration Overview

| Section  | Key                      | Type   | Default                 | Description                                            |
|----------|--------------------------|--------|-------------------------|--------------------------------------------------------|
| logging  | level                    | string | "info"                  | Log level: debug, info, warn, error.                   |
|          | format                   | string | "json"                  | Log format: json (production) or text (colorized dev). |
| server   | port                     | int    | 8080                    | The port on which the server runs.                     |
|          | app_url                  | string | "http://localhost:8080" | The base URL of the application.                       |
|          | allowed_origins          | array  | []                      | List of allowed CORS origins.                          |
| database | host                     | string | "localhost"             | Database host address.                                 |
|          | port                     | int    | 5432                    | Database port.                                         |
|          | user                     | string | "mrpix_sandbox"         | Database username.                                     |
|          | password                 | string | "sandXbox_mrpix2025"    | Database password.                                     |
|          | name                     | string | "mrpix_sandbox"         | Database name.                                         |
|          | sslmode                  | string | "disable"               | PostgreSQL SSL mode.                                   |
| auth     | jwt_secret               | string | "local-dev-secret"      | JWT signing secret.                                    |
|          | jwt_ttl_minutes          | int    | 480                     | JWT token TTL in minutes.                              |
|          | guest_jwt_ttl_minutes    | int    | 43200                   | Guest JWT token TTL in minutes.                        |
|          | guest_cookie_name        | string | "shopshredder_guest"    | Cookie name for guest tokens.                          |
| sandbox  | url_prefix               | string | "sandbox-"              | Prefix for sandbox URLs.                               |
|          | url_suffix               | string | ".localhost"            | Suffix for sandbox URLs.                               |
|          | default_lifetime         | int    | 60                      | Default sandbox lifetime in minutes.                   |
|          | max_lifetime             | int    | 1440                    | Maximum sandbox lifetime in minutes.                   |
|          | cleanup_interval_seconds | int    | 60                      | Interval between cleanup runs in seconds.              |
|          | internal_port            | int    | 80                      | Internal container port.                               |
| docker   | mode                     | string | "port"                  | Docker mode ("port" or "traefik").                     |
|          | network                  | string | "internal"              | Docker network name.                                   |
|          | traefik_enable           | bool   | false                   | Enable Traefik integration.                            |
|          | traefik_entrypoints      | string | "websecure"             | Traefik entrypoints.                                   |
|          | traefik_certresolver     | string | "production"            | Traefik certificate resolver.                          |
|          | traefik_middlewares      | string | ""                      | Traefik middlewares (comma-separated).                 |
|          | snapshot_author          | string | "shopshredder-api"      | Author for container snapshots.                        |
|          | snapshot_comment         | string | ""                      | Comment for container snapshots.                       |
| storage  | thumbnail_dir            | string | "storage/thumbnails"    | Directory for image thumbnail files.                   |
| guard    | max_total_sandboxes      | int    | 32                      | Maximum number of sandboxes allowed in total.          |
|          | max_sandboxes_per_ip     | int    | 5                       | Maximum number of concurrent sandboxes per IP address. |
|          | max_sandboxes_per_user   | int    | 10                      | Maximum number of concurrent sandboxes per user.       |

### Overriding Configuration with Environment Variables

The application supports environment variables to override configuration values.

| Environment Variable         | Corresponding Config Key     |
|------------------------------|------------------------------|
| LOGGING_LEVEL                | logging.level                |
| LOGGING_FORMAT               | logging.format               |
| SERVER_PORT                  | server.port                  |
| SERVER_APP_URL               | server.app_url               |
| DATABASE_HOST                | database.host                |
| DATABASE_PORT                | database.port                |
| DATABASE_USER                | database.user                |
| DATABASE_PASSWORD            | database.password            |
| DATABASE_NAME                | database.name                |
| DATABASE_SSLMODE             | database.sslmode             |
| AUTH_JWT_SECRET              | auth.jwt_secret              |
| AUTH_JWT_TTL_MINUTES         | auth.jwt_ttl_minutes         |
| SANDBOX_URL_PREFIX           | sandbox.url_prefix           |
| SANDBOX_URL_SUFFIX           | sandbox.url_suffix           |
| SANDBOX_DEFAULT_LIFETIME     | sandbox.default_lifetime     |
| SANDBOX_MAX_LIFETIME         | sandbox.max_lifetime         |
| DOCKER_MODE                  | docker.mode                  |
| DOCKER_NETWORK               | docker.network               |
| DOCKER_TRAEFIK_ENABLE        | docker.traefik_enable        |
| GUARD_MAX_TOTAL_SANDBOXES    | guard.max_total_sandboxes    |
| GUARD_MAX_SANDBOXES_PER_IP   | guard.max_sandboxes_per_ip   |
| GUARD_MAX_SANDBOXES_PER_USER | guard.max_sandboxes_per_user |

### Docker Compose Environment (`.env`)

Copied from `.env.example`. Controls Docker and database settings:

| Variable             | Description                                    | Default                 |
|----------------------|------------------------------------------------|-------------------------|
| `DOCKER_SOCKET_PATH` | Path to Docker socket (differs macOS vs Linux) | `/var/run/docker.sock`  |
| `POSTGRES_USER`      | PostgreSQL username                            | `mrpix_sandbox`         |
| `POSTGRES_PASSWORD`  | PostgreSQL password                            | /                       |
| `POSTGRES_DB`        | PostgreSQL database name                       | `mrpix_sandbox`         |
| `WEB_API_URL`        | API URL for web frontend                       | `http://localhost:8080` |

## Development

This project uses automated CI/CD pipeline and pushes production ready images to Github Packages.
The configuration and build options can be found at [.github/workflows/release.yml](.github/workflows/release.yml)

### API (`/api/`)

```bash
bab api:dev
```

Or manually:

```bash
cd api && go run ./cmd/api
```

### Database Migrations

Database schema changes are managed with [Goose](https://github.com/pressly/goose) using the SQL files in `api/internal/database/migrations`.

Use Bab to work with migrations:

```bash
bab db:migrate
bab db:migrate:status
bab db:migrate:create
bab db:migrate:fresh
```

Notes:

- `bab db:migrate` applies only pending migrations and records them in Goose's `goose_db_version` table.
- `bab db:migrate:create` creates a new sequential SQL migration with Goose annotations.
- `bab db:migrate:fresh` is destructive and intended for local development only.
- The production API image includes the `migrate` CLI and runs `migrate up` automatically on container start by default.
- Set `RUN_MIGRATIONS_ON_START=false` if migrations should be handled separately during deployment.

You can also run the project-local migration CLI directly:

```bash
cd api
go run ./cmd/migrate up
go run ./cmd/migrate status
go run ./cmd/migrate create add_user_index
```

#### Docker Host (macOS)

If using Docker Desktop on macOS, set your socket path in `.env`:

```
DOCKER_SOCKET_PATH=/Users/<username>/.docker/run/docker.sock
```
```
DOCKER_HOST=/Users/<username>/.docker/run/docker.sock
```

### Web Frontend (`/web/`)

```bash
bab web:dev
```

Or manually:

```bash
cp web/.env.example web/.env  # set WEB_API_URL
cd web && pnpm install && pnpm dev
```

### Custom Shopware Image Build

In order to provide a Shopware sandbox image, you can not use the default dockware image because it has missing
configuration options behind proxies.
To avoid getting Mixed-Content and HSTS problems in your browser, you have to extend the dockware image and build a
custom sandbox image.
Edit the Dockerfile in docker/images/.../Dockerfile and change the Shopware version in the FROM statement.
In addition you can add your custom configuration files to the image by copying them into the image.

Then, you can build your new image.

```bash
docker build . -t mr-pixel/sw-sandbox:6.7.0.0-rc1
```

After that, go to the administration page in the application and add this new image that you just created.
Now we have a plain shopware image that can be used for demo and testing purposes.

#### Custom Configuration

However, in order to create demonstration images with custom configurations, you have to first start a new sandbox and
choose your base image.
Then you can open your sandbox in the browser and install plugins, custom themes and configure the entire store as you
like.

After that you can create a new snapshot of the running sandbox container. Enter your image name and tag.
Congrats, you have created a custom Shopware demo sandbox image. Don't forget to stop the running sandbox container.

## Bab Tasks

This project uses [bab](https://bab.sh) as a task runner. Available tasks:

| Command              | Description                                  |
|----------------------|----------------------------------------------|
| `bab setup`          | Copy example configs for initial setup       |
| `bab dev`            | Start infrastructure + API + web dev servers |
| `bab api:dev`        | Run API locally                              |
| `bab web:dev`        | Run web frontend dev server                  |
| `bab docker:up`      | Start full stack                             |
| `bab docker:infra`   | Start infrastructure only (database)         |
| `bab docker:down`    | Stop all docker services                     |
| `bab docker:destroy` | Stop services and remove volumes             |
| `bab docker:logs`    | Follow logs from all services                |
| `bab docker:status`  | Show running services                        |
| `bab docker:build`   | Build all docker images                      |
