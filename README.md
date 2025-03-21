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






# Configuration Guide

This document provides an overview of the configuration settings for the application.  
The configuration is loaded from a `config.yml` file and can be overridden using environment variables.

## Configuration File (`config.yml`)

The application configuration is structured as follows:

```yaml
server:
  port: 8080
  allowed_origins:
    - "http://localhost:5173"
    - "https://www.shopshredder.de"
    - "http://localhost:8000"
    - "http://localhost:80"
  app_url: "http://localhost"

sandbox:
  url_prefix: "sandbox-"
  url_suffix: ".shopshredder.de"
  default_lifetime: 60

auth:
  username: "admin"
  password: "password"

guard:
  max_total_sandboxes: 1
  max_sandboxes_per_ip: 5
  max_sandbox_lifetime: 32

#database:
#  host: "localhost"
#  port: 5432
#  user: "postgres"
#  password: "password"
#  name: "appdb"
```


## Configuration Overview

| Section   | 	Key                  | 	Type    | 	Default            | 	Description                                            |
|-----------|-----------------------|----------|---------------------|---------------------------------------------------------|
| server    | 	port                 | 	int	    | 8080                | 	The port on which the server runs.                     |
|           | allowed_origins       | 	array   | 	[]                 | 	List of allowed CORS origins.                          |
|           | app_url               | 	string  | 	"http://localhost" | 	The base URL of the application.                       |
| sandbox   | 	url_prefix           | 	string  | 	"sandbox-"         | 	Prefix for sandbox URLs.                               |
|           | url_suffix            | 	string  | 	".shopshredder.de" | 	Suffix for sandbox URLs.                               |
|           | default_lifetime      | 	int	    | 60                  | 	Default sandbox lifetime in minutes.                   |
| auth      | 	username             | 	string  | 	"admin"            | 	Username for basic authentication.                     |
|           | password              | 	string	 | "password"          | 	Password for basic authentication.                     |
| guard     | 	max_total_sandboxes  | 	int	    | 32                  | 	Maximum number of sandboxes allowed in total.          |
|           | max_sandboxes_per_ip  | 	int	    | 5                   | 	Maximum number of concurrent sandboxes per IP address. |
|           | max_sandbox_lifetime	 | int	     | 60                  | 	Maximum sandbox lifetime in minutes.                   |
| database* | 	host                 | 	string  | 	"localhost"        | 	Database host address.                                 |
|           | port                  | 	int     | 	5432               | 	Database port.                                         |
|           | user                  | 	string  | 	"postgres"         | 	Database username.                                     |
|           | password              | 	string  | 	"password"         | 	Database password.                                     |
|           | name                  | 	string  | 	"appdb"            | 	Database name.                                         |

*Database configuration is currently disabled. Uncomment it in config.yml to enable.
If no database is specified a SQLite database is used automatically.

## Overriding Configuration with Environment Variables
The application supports environment variables to override configuration values.
Here is a list of environment variables and their corresponding settings:

| Environment Variable        | Corresponding Config Key   |
|-----------------------------|----------------------------|
| SERVER_PORT	                | server.port                |
| SANDBOX_URL_PREFIX          | sandbox.url_prefix         |
| SANDBOX_URL_SUFFIX          | sandbox.url_suffix         |
| SANDBOX_DEFAULT_LIFETIME    | sandbox.default_lifetime   |
| AUTH_USERNAME               | auth.username              |
| AUTH_PASSWORD               | auth.password              |
| GUARD_MAX_TOTAL_SANDBOXES   | guard.max_total_sandboxes  |
| GUARD_MAX_SANDBOXES_PER_IP	 | guard.max_sandboxes_per_ip |
| GUARD_MAX_SANDBOX_LIFETIME	 | guard.max_sandbox_lifetime |





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
