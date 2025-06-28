# Changelog

All notable changes to this project will be documented in this file.

## [5.0.0] - 2025-06-28
### Added
- Support for Windows Credential Manager via wincred
- Linux backend using libsecret, tested on Ubuntu 24.04, Amazon Linux 2 and RHEL 9
- `--stdout` option for `conceal get` to allow piping secrets
- Comprehensive documentation under `docs/`
- GitHub Actions workflows for linting, testing and releasing across platforms
- GoReleaser config and release workflow

### Changed
- Clipboard functions and keychain backends now fully testable
- Simplified README for newcomers and reorganized docs

## [4.0.0] - 2024-05-31
### Changed
- Updated conceal package version to 4.0.0

## [3.0.0] - 2020-12-01
### Added
- Windows release build and macOS executable

## [2.0.1] - 2020-07-09
### Fixed
- Debug message when retrieving secrets

## [2.0.0] - 2020-07-08
### Changed
- Removed godocdown dependency

## [1.0.4] - 2019-??
### Changed
- Version bump to 1.0.4-dev

## [1.0.3] - 2019-??
### Changed
- README updates

## [1.0.2] - 2019-??
### Changed
- Version bump

## [1.0.1] - 2019-??
### Changed
- Version bump

## [1.0.0] - 2019-??
### Added
- Initial release
