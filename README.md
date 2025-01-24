# mpXsandbox

This application allows you to create demo shops in a docker environment.


Start application with docker compose:
```
docker-compose up --build
```

## Backend

Requires a environment variable for docker and it has to be started as root:
```
DOCKER_HOST=unix:///Users/manuel.kienlein/.docker/run/docker.sock
```

### Swagger Docs
Command to compile swagger documentation page under http://localhost:8080/swagger/index.html
```
swag init
```


## Frontend
Start development server
```
yarn dev
```

# Old Documentation from Fork

This environment is in use internally for testing store plugins.

Each created instance has an own subdomain. The Shopware installation runs in a subfolder `/shop/public`.
The Adminer Plugin and App-System are preinstalled.

**This Application has only an API**

**This Application should run only in internal networks**

## Just running the Docker Container

```bash
docker run --rm -p 80:80 -e VIRTUAL_HOST=localhost ghcr.io/shopwarelabs/testenv:6.4.3
```

Access shop at http://localhost/shop/public

### Admin Credentials

User: `demo`
Password: `demodemo`

## API

### GET /environments

Returns all running containers


### POST /environments

JSON Request:

```json
{
    "installVersion": "<lowest supported version of plugin>",
    "plugin": "<plugin zip encoded>"
}
```

Response

```json
{
    "id": "<docker id>",
    "domain": "<running url>",
    "installedVersion": "<installed version>"
}
```

### DELETE /environments?id=dockerId

Response

```json
{
    "success": true
}
```
