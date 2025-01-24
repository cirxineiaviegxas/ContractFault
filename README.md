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
contractfault seismic report
service : orders-api
shift   : 1.4.0 -> 2.0.0
magnitude 8.5  verdict RUPTURE
[##################################------]
breaking=8  additive=5  behavioral=2  affected-consumers=4

CHANGES (sorted by severity)
  !! [critical] cancelOrder                  method changed POST -> DELETE
      code: endpoint.method.changed
      hits: checkout-web
      fix : Update the HTTP verb from POST to DELETE.
  !! [critical] getReceipt                   endpoint GET /orders/{id}/receipt (getReceipt) was removed
      code: endpoint.removed
      hits: analytics-etl, receipts-mailer
      fix : Stop calling this endpoint; migrate to a replacement operation before upgrading.
  ~~ [major   ] cancelOrder                  idempotency changed true -> false
      code: endpoint.idempotency.changed
      hits: checkout-web
      fix : Review retry logic: repeated calls may no longer be safe.
  !! [major   ] cancelOrder                  path changed /orders/{id}/cancel -> /orders/{id}
      code: endpoint.path.changed
      hits: checkout-web
      fix : Update the request URL template to /orders/{id}.
  !! [major   ] Order.couponCode             field Order.couponCode removed
      code: field.removed
      hits: analytics-etl, checkout-web
      fix : Stop reading Order.couponCode; it is no longer present.
  !! [major   ] listOrders#query:limit       parameter query:limit became required
      code: param.required.added
      hits: analytics-etl, fulfillment-worker
      fix : Always supply limit; requests without it will be rejected.
  !! [major   ] OrderStatus                  type OrderStatus dropped enum values refunded
      code: type.enum.removed
      fix : Handle the removed enum values as invalid; they will no longer be produced.
  ~~ [minor   ] OrderStatus                  type OrderStatus gained enum values authorized,delivered
      code: type.enum.added
      fix : Ensure consumers tolerate the new enum values.
  ++ [info    ] Money.display                field Money.display added (required=false)
      code: field.added
      hits: receipts-mailer
      fix : New optional field; read it when useful.

CONSUMER BLAST RADIUS
  checkout-web       high    score=35.3  breaking=3 additive=1 behavioral=1
  analytics-etl      medium  score=18.5  breaking=3 additive=1 behavioral=0
  fulfillment-worker high    score=8.3  breaking=1 additive=1 behavioral=0
  receipts-mailer    low     score=4.3  breaking=1 additive=1 behavioral=0
```

(The listing above is trimmed to the highest-severity tremors; the tool prints
all fifteen. The `##...` bar is the magnitude meter on a 0–10 scale.)

Read it top-down like a seismologist reads a drum: the biggest tremors are at
the top, each annotated with **who feels it** (`hits:`) and **how to survive it**
(`fix:`). The blast-radius footer ranks the towns by how hard they were shaken.

---

## The animated seismograph viewer

Pipe the JSON report into the TypeScript viewer to draw the fault as a station
map. Each consumer becomes a seismic station along the fault line; its dot grows
with its risk score, and its color encodes the worst tremor it felt.

```bash
# Emit JSON from the analyzer, then render an SVG seismograph
./bin/contractfault -old examples/contracts/orders-v1.json \
    -new examples/contracts/orders-v2.json \
    -consumers "examples/consumers/*.json" \
    -format json -out examples/report.json

cd viewer
node dist/cli.js ../examples/report.json --svg ../docs/assets/impact-seismograph.svg
```

<div align="center">
<img src="docs/assets/impact-seismograph.svg" alt="impact seismograph" width="80%" />
</div>

The viewer also renders a colorized terminal map with per-consumer amplitude
bars, and — like the Go CLI — exits `2` on breaking, `1` on behavioral, `0` on
stable, so it can double as a CI gate when reading a stored report.

---

## What counts as a tremor

contractfault classifies every difference into one of three categories, and each
category maps to a CI outcome. This is the crux of *not* being a text diff: the
tool understands the **semantics** of a contract, not its bytes.

| Category       | Feels like      | Examples                                              | CI |
|----------------|-----------------|-------------------------------------------------------|----|
| **breaking**   | the ground cracks | endpoint removed, method flip, required field added, enum value dropped, response type changed | exit 2 |
| **behavioral** | the ground shifts | idempotency flip, field became nullable, enum value *added*, field deprecated | exit 1 |
| **additive**   | the ground settles | new optional field, new endpoint, new status code, parameter relaxed to optional | exit 0 |

The subtle cases are where the design earns its keep:

- **A required field added is breaking; an optional one is additive.** Same JSON
  edit, opposite blast radius.
- **Adding an enum value is *behavioral*, not additive** — old consumers with an
  exhaustive `switch` may fall through on the new value.
- **An idempotency flip is behavioral even though the shape is byte-identical** —
  it silently invalidates retry logic. A text differ would never catch this.
- **A parameter *addition* shakes every caller of the endpoint**, but a
  parameter *value* change only shakes consumers that set that specific param.

The complete rule table (thirty-plus codes) lives in
[`docs/CONTRACT.md`](docs/CONTRACT.md#33-rule-table).

---

## How the blast radius is computed

A tremor with no town on top of it is a curiosity, not an incident.
contractfault only counts a change against a consumer when that consumer
actually depends on the affected element:

