# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).


## [1.4.3] - 2026-04-08
### Added
- pgadmin service to docker stack for production

### Fixed
- Terminal hijack didn't work

## [1.4.2] - 2026-04-08
### Fixed
- Shredder animation always playing
- Thumbnail fit in sandbox card

## [1.4.1] - 2026-04-08
### Fixed
- Registry default value didn't get used in backend logic
- Swagger docs folder not needed anymore in Dockerfile

## [1.4.0] - 2026-04-08
### Added
- Public sandbox endpoints for guests
- Pagination support (API and frontend)

### Changed
- Migrated to Fuego REST API framework
- Refactored auth to stateless JWT (instead of session-based)
- Renamed `guest_session_id` to `client_id`
- REST best practices and consistent handler/DTO/service patterns
- Updated frontend types and DTOs for new API endpoints

### Removed
- Legacy REST API files

### Fixed
- Vertical cropping of thumbnails in preset and sandbox cards
- Wrong migration format

## [1.3.3] - 2026-04-02
### Fixed
- Missing metadata enrichment

## [1.3.2] - 2026-04-02
### Changed
- Health checks disabled by default

## [1.3.1] - 2026-04-01
### Added
- Audit log query parameters, facets, and detail view
- Admin guards on GET sandboxes and audit-logs endpoints
- Unit and integration tests for audit log
- Audit log documentation

### Changed
- Refactored audit log system with query parameters, meta/data fields, facets, and detail view
- Renamed AuditLog `createdAt` to `timestamp`

### Fixed
- Static list of audit log actions in frontend
- Only log sandbox delete after successful container deletion
- Missing environment mapping in docker compose
- Missing origin header
- Release workflow and load balancer Traefik label
- Docker network SSH stream (primary network with localhost fallback)

## [1.3.0] - 2026-04-01
### Added
- Sandbox details dialog with SSH credentials
- Shell terminal integrated as tab in sandbox details (via WebSocket docker exec)
- Real-time sandbox observation via SSE endpoint
- SSH proxy support

### Changed
- Improved polling approach using SSE for sandboxes

### Fixed
- Missing health endpoint
- Private images not shown in create sandbox dialog
- Full height image list in create sandbox dialog
- Missing owner object

## [1.2.0] - 2026-03-31
### Added
- Sandboxes with unlimited lifetime

### Changed
- Improved user interface and user experience

### Fixed
- Sandbox not visible in Gallery when not logged in (Guest-Mode) [#58](https://github.com/mr-pixel-kg/shopware-sandbox-plattform/issues/58)
- APP_URL not set correctly [#59](https://github.com/mr-pixel-kg/shopware-sandbox-plattform/issues/59)

## [1.1.0] - 2026-03-27
### Added
- User management
- Registration modes:
  - Open registration
  - Invite-only registration
- Image registry
  - Metadata for images
  - Custom action buttons
  - Credentials
  - Post start commands
  - Environment variables
- Display Names for Sandbox Instances

### Changed
- Improved user interface and user experience

## [1.0.0] - 2026-03-24
This is a complete rewrite of the project, with a new architecture and implementation.

### Added
- New api backend
- New web frontend
- JWT auth and guest sessions
- Public demo gallery
- Private sandboxes
- Sandbox health monitoring
- Snapshot image creation
- OpenAPI documentation
- Audit log
- Thumbnail upload
- Sandbox Image Management
- Sandbox Management

### Changed
- Complete rewrite
- New project architecture
- Reworked sandbox lifecycle
- Improved logging
- Improved error handling

### Removed
- Legacy implementation

## [0.3.0] - 2026-03-18
### Added
- Sentry integration
- Plugin templates

### Fixed
- Display of expiration time progressbar
- Darkmode colors
- Error messages

### Changed
- Improved logging

## [0.2.0] - 2025-04-02
### Added
- Audit Log
- Create snapshot

### Fixed
- Dark mode design
- Nginx configuration to redirect all requests to SPA index.

## [0.1.0] - 2025-03-26
### Added
- Basic administration
  - Management of sandbox environments
  - Management of sandbox images
- Basic storefront
  - Gallery of sandbox images
  - List of running sandbox environments
- Rate Limiting of max running sandboxes (per IP and total system)
- Auto-removal of sandboxes with expired lifetime
- Garbage Collector
- Startup check procedure
