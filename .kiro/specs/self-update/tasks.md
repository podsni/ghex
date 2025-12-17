# Implementation Plan

- [x] 1. Set up update package structure and dependencies


  - Create `internal/update/` directory structure
  - Add gopter dependency for property-based testing: `go get github.com/leanovate/gopter`
  - Create base error types in `internal/update/errors.go`
  - _Requirements: 2.6, 2.7_





- [ ] 2. Implement version parsing and comparison
  - [ ] 2.1 Create Version struct and parsing logic
    - Implement `ParseVersion()` to parse semantic version strings (e.g., "1.0.0", "v1.2.3-beta.1")
    - Implement `Version.String()` to convert back to string
    - Implement `Version.Compare()` for version comparison
    - Implement `Version.IsNewerThan()` helper method
    - _Requirements: 1.2_
  - [ ]* 2.2 Write property test for version comparison transitivity
    - **Property 1: Semantic Version Comparison Transitivity**




    - **Validates: Requirements 1.2**
  - [ ]* 2.3 Write property test for version parsing round-trip
    - **Property 6: Version Parsing Round-Trip**
    - **Validates: Requirements 1.2**

- [x] 3. Implement checksum verification




  - [ ] 3.1 Create checksum utilities
    - Implement `CalculateChecksum()` to compute SHA256 hash of a file
    - Implement `VerifyChecksum()` to verify file against expected hash
    - Implement `ParseChecksumFile()` to parse checksums.txt format
    - _Requirements: 2.3_
  - [ ]* 3.2 Write property test for checksum verification
    - **Property 3: Checksum Verification Correctness**
    - **Validates: Requirements 2.3**







- [ ] 4. Implement binary manager for backup/restore
  - [ ] 4.1 Create BinaryManager struct and methods
    - Implement `GetCurrentBinaryPath()` to get running binary location
    - Implement `BinaryManager.Backup()` to create backup before update




    - Implement `BinaryManager.Restore()` to restore from backup
    - Implement `BinaryManager.HasBackup()` to check backup existence
    - Implement `SetExecutable()` for Unix permission handling
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 4.1_
  - [ ]* 4.2 Write property test for backup/restore round-trip
    - **Property 5: Backup/Restore Round-Trip**




    - **Validates: Requirements 5.1, 5.2**

- [ ] 5. Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 6. Implement GitHub client for release fetching
  - [ ] 6.1 Create GitHubClient and release fetching
    - Implement `GitHubClient.GetLatestRelease()` to fetch latest release info



    - Implement `GitHubClient.GetReleases()` to fetch multiple releases for changelog

    - Implement `GitHubClient.DownloadAsset()` with progress callback
    - Parse GitHub API JSON response into ReleaseInfo struct
    - _Requirements: 1.1, 3.1, 3.3_

- [x] 7. Implement platform detection and asset selection

  - [ ] 7.1 Create platform-aware asset selection
    - Implement `GetAssetName()` to construct asset filename for current platform




    - Implement `SelectAsset()` to find matching asset from release
    - Handle all supported platforms: linux/darwin/windows with amd64/arm64
    - _Requirements: 2.1_


  - [ ]* 7.2 Write property test for platform URL construction
    - **Property 2: Platform Asset URL Construction**
    - **Validates: Requirements 2.1**

- [ ] 8. Implement core Updater logic
  - [ ] 8.1 Create Updater struct and update flow
    - Implement `NewUpdater()` constructor
    - Implement `Updater.CheckForUpdate()` to check if newer version exists



    - Implement `Updater.Update()` orchestrating: download → verify → backup → replace
    - Implement `Updater.Rollback()` to restore previous version
    - Implement `Updater.GetChangelog()` to fetch release notes between versions
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 2.1, 2.2, 2.3, 2.4, 2.5, 3.1, 3.2, 3.3_
  - [ ]* 8.2 Write property test for error handling preservation
    - **Property 4: Error Handling Preserves Installation**
    - **Validates: Requirements 2.6, 2.7**

- [ ] 9. Implement CLI update command
  - [ ] 9.1 Create update command with flags
    - Create `cmd/ghex/commands/update.go` with cobra command
    - Implement `--check` flag for checking updates only
    - Implement `--changelog` flag for showing release notes
    - Implement `--rollback` flag for restoring previous version
    - Implement `--force` and `--yes` flags for non-interactive mode
    - Add progress indicator during download using existing UI components
    - _Requirements: 1.1, 1.3, 2.2, 2.5, 3.2, 5.2_
  - [ ] 9.2 Register update command in root
    - Add `NewUpdateCmd()` to root command in `cmd/ghex/commands/root.go`
    - _Requirements: 1.1_

- [ ] 10. Implement permission handling
  - [ ] 10.1 Add permission detection and handling
    - Implement `CheckWritePermission()` to verify update is possible
    - Add platform-specific handling for Windows file replacement
    - Provide helpful error messages with instructions for elevated permissions
    - _Requirements: 4.1, 4.2, 4.3, 4.4_

- [ ] 11. Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [ ]* 12. Write integration tests
  - [ ]* 12.1 Create integration tests for update flow
    - Test complete update flow with mock GitHub API
    - Test rollback scenario
    - Test error recovery scenarios
    - _Requirements: 2.4, 2.5, 2.6, 2.7, 5.2, 5.3_

- [ ] 13. Final Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.
