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
| Web       | `/web/`   | Vue 3, Shadcn-vue, TypeScript, npm   |

## Configuration

### API Configuration (`api/config.yml`)

Copy `api/config.example.yml` to `api/config.yml` and edit with your settings. The example is pre-configured for local
development (localhost database, dev JWT secret, Traefik disabled).

Key sections:

- `server`: port, app URL, CORS origins
- `database`: PostgreSQL connection (matches docker-compose credentials)
- `auth`: JWT secret and token TTL
- `sandbox`: URL pattern, lifetime, cleanup interval
- `docker`: Docker/Traefik integration for sandbox containers
- `guard`: rate limits (max sandboxes per IP/user)

### Docker Compose Environment (`.env`)

Copied from `.env.example`. Controls Docker and database settings:

| Variable             | Description                                    | Default                 |
|----------------------|------------------------------------------------|-------------------------|
| `DOCKER_SOCKET_PATH` | Path to Docker socket (differs macOS vs Linux) | `/var/run/docker.sock`  |
| `POSTGRES_USER`      | PostgreSQL username                            | `mrpix_sandbox`         |
| `POSTGRES_PASSWORD`  | PostgreSQL password                            | /                       |
| `POSTGRES_DB`        | PostgreSQL database name                       | `mrpix_sandbox`         |
| `VITE_BACKEND_URL`   | Backend URL for frontend                       | `http://localhost:8080` |

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

#### Docker Host (macOS)

If using Docker Desktop on macOS, set your socket path in `.env`:

```
DOCKER_SOCKET_PATH=/Users/<username>/.docker/run/docker.sock
```

### Web Frontend (`/web/`)

```bash
bab web:dev
```

Or manually:

```bash
cp web/.env.example web/.env  # set VITE_API_BASE_URL
cd web && npm install && npm run dev
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
