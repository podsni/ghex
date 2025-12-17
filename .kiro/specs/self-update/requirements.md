# Requirements Document

## Introduction

Fitur Self-Update memungkinkan ghex CLI untuk mengupdate dirinya sendiri ke versi terbaru secara otomatis dari GitHub Releases. Fitur ini memberikan pengalaman yang seamless bagi pengguna untuk selalu menggunakan versi terbaru tanpa perlu mengunduh dan menginstall secara manual.

## Glossary

- **GHEX**: CLI tool untuk mengelola multiple GitHub accounts dan universal downloader
- **Self-Update**: Kemampuan aplikasi untuk mengupdate dirinya sendiri ke versi terbaru
- **GitHub Releases**: Fitur GitHub untuk mendistribusikan versi rilis software
- **Semantic Version**: Format versi menggunakan MAJOR.MINOR.PATCH (contoh: 1.0.0)
- **Binary**: File executable yang dapat dijalankan langsung
- **Checksum**: Hash untuk memverifikasi integritas file yang diunduh

## Requirements

### Requirement 1

**User Story:** As a ghex user, I want to check if a newer version is available, so that I can decide whether to update.

#### Acceptance Criteria

1. WHEN a user runs the update check command THEN the GHEX System SHALL fetch the latest release information from GitHub Releases API
2. WHEN the latest version is fetched THEN the GHEX System SHALL compare the current version with the latest version using semantic versioning
3. WHEN a newer version is available THEN the GHEX System SHALL display the current version, latest version, and release notes summary
4. WHEN the current version is already the latest THEN the GHEX System SHALL inform the user that no update is available

### Requirement 2

**User Story:** As a ghex user, I want to update ghex to the latest version with a single command, so that I can quickly get new features and bug fixes.

#### Acceptance Criteria

1. WHEN a user runs the update command THEN the GHEX System SHALL download the appropriate binary for the current platform and architecture
2. WHEN downloading the binary THEN the GHEX System SHALL display a progress indicator showing download status
3. WHEN the download completes THEN the GHEX System SHALL verify the downloaded file integrity using checksum
4. WHEN verification succeeds THEN the GHEX System SHALL replace the current binary with the new version
5. WHEN the update completes successfully THEN the GHEX System SHALL display the new version number and confirmation message
6. IF the download fails THEN the GHEX System SHALL display an error message and preserve the current installation
7. IF checksum verification fails THEN the GHEX System SHALL abort the update and display a security warning

### Requirement 3

**User Story:** As a ghex user, I want to see the changelog before updating, so that I can understand what changes are included.

#### Acceptance Criteria

1. WHEN a user requests changelog information THEN the GHEX System SHALL fetch and display the release notes from GitHub
2. WHEN displaying release notes THEN the GHEX System SHALL format the content for terminal readability
3. WHEN multiple versions are available THEN the GHEX System SHALL show changes between current and latest version

### Requirement 4

**User Story:** As a ghex user, I want the update process to handle permissions correctly, so that the update works on different operating systems.

#### Acceptance Criteria

1. WHEN updating on Unix systems THEN the GHEX System SHALL preserve executable permissions on the new binary
2. WHEN updating on Windows THEN the GHEX System SHALL handle file replacement considering running process limitations
3. IF insufficient permissions exist THEN the GHEX System SHALL provide clear instructions for manual update or elevated permissions
4. WHEN the binary location requires elevated permissions THEN the GHEX System SHALL detect this and inform the user before attempting update

### Requirement 5

**User Story:** As a ghex user, I want to rollback to the previous version if the update causes issues, so that I can maintain a working installation.

#### Acceptance Criteria

1. WHEN performing an update THEN the GHEX System SHALL create a backup of the current binary before replacement
2. WHEN a user requests rollback THEN the GHEX System SHALL restore the previous version from backup
3. WHEN backup restoration completes THEN the GHEX System SHALL verify the restored binary is functional
4. IF no backup exists THEN the GHEX System SHALL inform the user that rollback is not available

