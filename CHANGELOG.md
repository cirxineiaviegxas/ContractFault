# Changelog

All notable changes to contractfault are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and the project adheres
to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-08-31

### Added

- Go stdlib CLI `contractfault` that ingests two versions of a documented JSON
  contract plus consumer usage manifests and emits deterministic JSON/text
  seismic reports with CI exit codes (0 stable, 1 shaken, 2 rupture, 3 usage).
- Change classification engine covering endpoints, parameters, responses,
  reusable types, fields, enums, nullability, arity, required-ness,
  deprecation and idempotency — each mapped to breaking / additive / behavioral.
- Consumer blast-radius join: every change is attributed to the named consumers
