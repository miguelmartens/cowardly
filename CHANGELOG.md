# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Desired state and re-apply: config in `~/.config/cowardly/cowardly.yaml`, `--reapply`, `--install-login-hook`, TUI detection of reverted settings (press R to re-apply).
- Release assets are now `.tar.gz` archives containing `cowardly`, CHANGELOG.md, LICENSE, and README.md; asset names follow `cowardly_v{VERSION}_{OS}_{ARCH}.tar.gz`.
