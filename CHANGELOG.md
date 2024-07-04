# Changelog

All notable changes for every released version of this project will be documented in this file.  
Check [Changelog Format](#Changelog-Format) for more details about this changelog's format.

## Unreleased

### Changed

- Improve comments' content
- Removed unnecessary patch comment

## [v0.6.0] - 2024-04-26

### Added

- Add details and URLs of the generated artifacts by the GitHub Actions
- Append error details in the patch's comment in case of an error occurs
- Parsing CI Broker's message protocol version
- Support for CI Broker message protocol version 1

### Changed

- Support single comment in a patch (instead of multiple) which is updated periodically with workflows' progress and 
  results

### Removed

- Broker's Error Response which supported the Error variant (without protocol version support)

### Fixed

- Documentation issues and updates

## [v0.5.0] - 2024-03-21

### Added

- Configure through env vars
- Support multiple GitHub Action workflows monitoring
- Support add comments on patch with information for the GitHub Actions and results in Markdown format

## Changelog Format

Title: **Version ([vX.Y.Z]) - Date (YYYY-MM-DD)**

Version numbering follows the [Semantic Versioning](https://semver.org/spec/v2.0.0.html) schema.
You can expect the following sections under each version:

* Added: for new features.
* Changed: for changes in existing functionality.
* Deprecated: for soon-to-be removed features.
* Removed: for now removed features.
* Fixed: for any bug fixes.
* Security: in case of vulnerabilities.
