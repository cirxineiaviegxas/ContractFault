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
  that actually depend on the affected element, weighted by criticality.
- Seismic magnitude scoring on a compressed 0–10 scale with a plain-language
  verdict (`stable`, `tremor`, `shaken`, `rupture`).
- TypeScript seismic viewer that consumes the JSON report and renders both a
  colorized terminal impact map and a standalone animated SVG seismograph.
- Example before/after contracts for `orders-api` (v1.4.0 → v2.0.0) and four
  consumer manifests spanning low/medium/high criticality.
- Documentation: `docs/CONTRACT.md` specifying the contract and manifest format
  and the full classification table.
- Hand-authored animated hero SVG and a generated impact seismograph SVG under
  `docs/assets/`.
- Test suites for every Go package and the TypeScript viewer.
