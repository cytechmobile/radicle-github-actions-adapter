# Changelog

All notable changes for every released version of this project will be documented in this file.  
Check [Changelog Format](#Changelog-Format) for more details about this changelog's format.

## Unreleased

### Added

- Add details and URLs of the generated artifacts by the GitHub Actions
- Append error details in the patch's comment in case of an error occurs

### Changed

- Support single comment in a patch (instead of multiple) which is updated periodically with workflows' progress and 
  results

### Fixed

- Documentation issues and updates

## [0.5.0] - 2024-03-21

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
