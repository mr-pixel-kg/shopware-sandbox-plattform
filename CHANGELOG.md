# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
