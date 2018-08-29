# The contractfault contract & manifest format

This document specifies the two input formats that `contractfault` ingests —
the **contract** and the **consumer manifest** — and the full **change
classification table** the analyzer applies. Everything here is enforced by the
Go loaders (`internal/contract`, `internal/consumer`) and the analysis engine
(`internal/analyze`); the format is intentionally small enough to author by
hand yet expressive enough to detect consumer-affecting change.

---

## 1. Contract document

A contract is a single JSON object describing one version of an API surface.

```json
{
  "service": "orders-api",
  "version": "1.4.0",
