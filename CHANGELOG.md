# Changelog

All notable changes to this project will be documented in this file.

## [1.0.0] - 2025-10-15

### Added
- Created a Go application to download attachments and metadata from the ZenGRC API.
- Implemented a worker pool for concurrent processing of records.
- Added command-line flags for configuration: `-api-url`, `-token`, `-output-dir`, `-workers`, and `-overwrite`.
- Created a dedicated API client in `client.go` to handle all interactions with the ZenGRC API.
- Expanded data models in `client.go` to be fully compliant with the API specification.
- Implemented robust error handling using an error channel to collect and report errors from workers.
- Added a `CHANGELOG.md` file to document changes.
- Added comprehensive `README.md` with sections on overview, architecture, and usage examples.
- Added inline documentation to `main.go` and `client.go` for better code clarity.

### Changed
- Set secure file permissions (`0755` for directories, `0644` for files) to prevent unauthorized access.
- Centralized all API endpoint paths as constants in `client.go` for better maintainability.
- Optimized the HTTP client with a custom transport for better performance and connection pooling.

### Fixed
- Corrected an issue where the `strings` package was imported but not used.
- Resolved an issue where `main.go` and `client.go` were not compiled together, causing `undefined` errors.