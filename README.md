<!-- contractfault README — a seismic fault laboratory -->

<div align="center">

<img src="docs/assets/faultline-hero.svg" alt="contractfault faultline hero" width="100%" />

# contractfault

**An API contract consumer-impact analyzer that treats every version bump like a seismic event.**

*Ingest two contract versions. Feel where the ground moves. See which downstream towns get shaken.*

</div>

---

## Why this exists — the seismology of APIs

Most "API diff" tools hand you a wall of text and let you guess what matters.
That is like handing a geologist a list of every rock that moved and asking them
to intuit whether the village downhill is in danger. **contractfault is not a
text differ.** It is a seismograph for API contracts.

The mental model is deliberately geological:

- A **contract** is a section of crust — a versioned map of endpoints and types.
- A **change** between two versions is a **tremor** along a fault line.
- Each tremor has a **category** (does the ground crack, shift, or just settle?)
  and a **severity** (how much energy was released?).
- **Consumers** are the towns built along the fault. A tremor only matters if a
  town sits on top of it.
- The aggregate energy, felt through the towns, is the **magnitude** — a single
  0–10 number with a plain-language **verdict**: `stable`, `tremor`, `shaken`,
  or `rupture`.

That framing is not decoration. It is the actual data model. contractfault
refuses to tell you "47 things changed." It tells you *"`checkout-web` (high
criticality) will rupture because `Order.couponCode` — which it reads — was
removed, and `cancelOrder` — which it calls — changed from POST to DELETE."*
That sentence is worth more than any diff.

```
                    v1.4.0  ──────────────────────────  v2.0.0
                       │                                    │
        ┌──────────────┴───────────┐          ┌─────────────┴──────────────┐
        │  endpoints · types        │  DIFF   │  endpoints · types          │
        └──────────────┬───────────┘   ═══>   └─────────────┬──────────────┘
                       │                                    │
                       ▼                                    ▼
                  ╭─────────────────  classify  ─────────────────╮
                  │  breaking · additive · behavioral            │
                  ╰───────────────────┬──────────────────────────╯
                                      │  join
                       ┌──────────────┼───────────────┐
                       ▼              ▼               ▼
                 checkout-web   analytics-etl   receipts-mailer   … named consumers
```

---

## The laboratory at a glance

```
contractfault/
├── cmd/contractfault/        # Go stdlib CLI — the seismograph
│   └── main.go               #   flag parsing, glob expansion, pipeline, exit codes
├── internal/
│   ├── contract/             # the documented JSON contract format + strict loader
│   ├── consumer/             # consumer usage manifests + criticality weighting
│   ├── analyze/              # the classification engine (the fault mechanics)
│   ├── impact/               # change→consumer join, magnitude, verdict
│   └── report/               # deterministic JSON + text seismograph renderers
├── viewer/                   # TypeScript seismic impact map / SVG viewer
│   ├── src/{types,seismic,cli}.ts
│   └── test/viewer.test.ts
├── examples/
│   ├── contracts/orders-v{1,2}.json    # before / after
│   ├── consumers/*.json                # four named consumers
│   └── report.json                     # a generated report fixture
├── docs/
│   ├── CONTRACT.md           # the format spec + full classification table
│   └── assets/               # animated faultline-hero + impact-seismograph SVGs
├── go.mod · Makefile · LICENSE · CHANGELOG.md · .github/workflows/ci.yml
```

Two languages, one contract. The Go analyzer is the instrument; the TypeScript
viewer is the paper the trace is drawn on. They meet at a single JSON schema
(`contractfault/v1`), so either half can be replaced without touching the other.

---

## Quick start — trigger your first quake

You need Go 1.24+ and Node 20+.

```bash
# 1. Build both halves
make build

# 2. Run the analyzer over the shipped example fault
./bin/contractfault \
    -old examples/contracts/orders-v1.json \
    -new examples/contracts/orders-v2.json \
    -consumers "examples/consumers/*.json"

# 3. Or run the whole pipeline — analyze, then render the seismic map + SVG
make demo
```

No flags to memorize: `-old`, `-new`, and a `-consumers` glob are the whole
interface. Everything else has a sensible default.

---

## The terminal seismograph (real output)

Running the analyzer against the bundled `orders-api` fault produces this. It is
copied verbatim from an actual run — no artistic license:

```text
