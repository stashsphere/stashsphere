# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- public links: share things and lists with users not registered on an instance
- ability to export and import all data, useful when changing instances and backups
- added schema based properties with dynamic, validating editors
- added the boolean property type
- backend: expose the reason why a thing / list is visible to a user
- backend: add `orderBy` and `orderDirection` to `/api/things`, allows to display directly shared things first
- backend: add `orderBy` and `orderDirection` to `/api/lists`, allows to display directly shared lists first
- add defaultSharingState to profile, creates things and lists with that sharing setting
- development: a nix-based development environment is available
  with the `dev` flake output.
- backend: add OIDC auth mechanism
  multiple providers can now be added for external authentication
- frontend: allow to paste images from clipboard
- frontend: allow to drag images from other applications / desktop
- frontend: add more icons for known properties

### Changed

- backend: all config values may now be set by environment variables as well:
  nesting is expressed as `__` while a single `_` is converted later to camelCase (`yaml`-like)
  Example: `STASHSPHERE_AUTH__PRIVATE_KEY` -> `AUTH.PRIVATE_KEY` -> `auth.privateKey`
  the existing value `STASHSPHERE_DISABLE_SECURE_COOKIES` must be changed to `STASHSPHERE_AUTH__DISABLE_SECURE_COOKIES`
- backend: `domains.api` config is deprecated, use `domains.cookieDomain` instead

### Fixed

- frontend: datalist ids for suggestions in PropertyEditor are now unique across rows

### Removed

## [0.9.0] - 2026-02-16

First tag / release.
