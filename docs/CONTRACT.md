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
  "types": { "...": { } },
  "endpoints": [ { } ]
}
```

| Field       | Type                | Required | Meaning                                             |
|-------------|---------------------|----------|-----------------------------------------------------|
| `service`   | string              | yes      | Logical API name. Must match across compared files. |
| `version`   | string              | yes      | Human-readable version, e.g. `2.0.0`.               |
| `types`     | object of `Type`    | no       | Reusable named types keyed by type name.            |
| `endpoints` | array of `Endpoint` | yes      | Operations exposed by the service.                  |

The loader uses strict decoding (`DisallowUnknownFields`): any unrecognized key
is a hard error, so typos surface immediately rather than being silently
ignored.

### 1.1 Type

```json
"Order": {
  "kind": "object",
  "fields": {
    "id":     { "type": "string", "required": true },
    "status": { "type": "OrderStatus", "required": true }
  }
}
```

| Field        | Type                | Meaning                                                        |
|--------------|---------------------|----------------------------------------------------------------|
| `kind`       | string              | `object`, `string`, `integer`, `number`, or `boolean`.         |
| `fields`     | object of `Field`   | Members of an `object` type.                                   |
| `enum`       | array of string     | Permitted values for a scalar type (empty = unconstrained).    |
| `deprecated` | boolean             | Marks the whole type as scheduled for removal.                 |

### 1.2 Field

| Field        | Type    | Meaning                                                                    |
|--------------|---------|----------------------------------------------------------------------------|
| `type`       | string  | Builtin scalar or the name of another `Type`.                              |
| `required`   | boolean | Field must always be present.                                              |
| `nullable`   | boolean | Field may carry an explicit `null`.                                        |
| `array`      | boolean | Field is a homogeneous list of `type`.                                     |
| `deprecated` | boolean | Field scheduled for removal.                                               |
| `doc`        | string  | Short description, surfaced verbatim in migration hints.                   |

### 1.3 Endpoint

```json
{
  "id": "getOrder",
  "method": "GET",
  "path": "/orders/{id}",
  "idempotent": true,
  "params": [ { "name": "id", "in": "path", "type": "string", "required": true } ],
  "responses": { "200": "Order", "404": "Error" }
}
```

| Field         | Type              | Meaning                                                                 |
|---------------|-------------------|-------------------------------------------------------------------------|
| `id`          | string            | **Stable** operation id — the correlation key across versions.          |
| `method`      | string            | HTTP verb, upper-cased on load.                                         |
| `path`        | string            | Templated URL path.                                                     |
| `params`      | array of `Param`  | Path / query / header parameters.                                       |
| `requestType` | string            | Name of the request body `Type` (omit for bodiless operations).         |
| `responses`   | object            | Maps status code (string) → response body `Type` name.                  |
| `deprecated`  | boolean           | Operation scheduled for removal.                                        |
| `idempotent`  | boolean           | Whether repeated calls are safe (a change here is *behavioral*).        |

Because `id` is the correlation key, you may freely rename a `path` or change a
`method` and contractfault still recognizes it as the *same* operation and
reports the mutation, rather than a delete + add.

### 1.4 Param

| Field      | Type            | Meaning                                    |
|------------|-----------------|--------------------------------------------|
| `name`     | string          | Parameter name.                            |
| `in`       | string          | `path`, `query`, or `header`.              |
| `type`     | string          | Scalar type of the value.                  |
| `required` | boolean         | Parameter must be supplied.                |
| `enum`     | array of string | Permitted values.                          |

A parameter's correlation key is `"{in}:{name}"`, e.g. `query:status`.

---

## 2. Consumer manifest

A manifest declares which parts of the contract a named downstream service
actually depends on, so impacts can be attributed precisely.

```json
{
  "name": "checkout-web",
  "team": "storefront",
  "criticality": "high",
  "uses": [
    {
      "endpoint": "createOrder",
      "writesFields": ["CreateOrderRequest.items"],
      "readsFields": ["Order.id", "Order.total"],
      "params": ["query:status"]
    }
  ]
}
```

| Field         | Type            | Meaning                                                             |
|---------------|-----------------|---------------------------------------------------------------------|
| `name`        | string          | Consumer service name (unique across the manifest set).             |
| `team`        | string          | Owning team, surfaced for routing.                                  |
| `criticality` | string          | `low`, `medium`, or `high` — weights the magnitude score (1/2/3).   |
| `uses`        | array of `Usage`| Dependency edges into the contract.                                 |

### 2.1 Usage

| Field          | Type            | Meaning                                                       |
|----------------|-----------------|---------------------------------------------------------------|
| `endpoint`     | string          | Operation id the consumer calls.                              |
| `readsFields`  | array of string | `Type.field` paths read from responses.                       |
| `writesFields` | array of string | `Type.field` paths sent in requests.                          |
| `params`       | array of string | Parameter keys (`query:status`) the consumer sets.            |

---

## 3. Change classification

Every difference between two versions is classified into a **category** and a
**severity**. Categories drive the CI exit code; severities drive ordering and
the seismic magnitude.

### 3.1 Categories

- **breaking** — can cause existing consumers to fail (exit code 2).
- **behavioral** — same shape, different runtime semantics (exit code 1).
- **additive** — extends the surface without breaking callers (exit code 0).

### 3.2 Severities

`critical` > `major` > `minor` > `info`, with magnitude weights
`4.0 / 2.5 / 1.0 / 0.25`.

### 3.3 Rule table

| Code                            | Category    | Severity | Trigger                                              |
|---------------------------------|-------------|----------|------------------------------------------------------|
| `endpoint.removed`              | breaking    | critical | An operation id disappears.                          |
| `endpoint.method.changed`       | breaking    | critical | HTTP verb changes for an id.                         |
| `endpoint.path.changed`         | breaking    | major    | Templated path changes for an id.                    |
| `endpoint.requestType.changed`  | breaking    | major    | Request body type changes.                           |
| `endpoint.added`                | additive    | info     | A new operation id appears.                          |
| `endpoint.deprecated`           | behavioral  | minor    | Operation newly marked deprecated.                   |
| `endpoint.idempotency.changed`  | behavioral  | major    | Idempotency flag flips.                              |
| `param.added` (required)        | breaking    | major    | A new required parameter.                            |
| `param.added` (optional)        | additive    | info     | A new optional parameter.                            |
| `param.removed` (was required)  | breaking    | major    | A required parameter is dropped.                     |
| `param.removed` (was optional)  | additive    | minor    | An optional parameter is dropped.                    |
| `param.required.added`          | breaking    | major    | Existing parameter becomes required.                 |
| `param.required.relaxed`        | additive    | info     | Existing parameter becomes optional.                 |
| `param.enum.removed`            | breaking    | major    | A parameter drops accepted enum values.              |
| `response.removed`              | breaking    | major    | A documented status code disappears.                 |
| `response.type.changed`         | breaking    | major    | A status code's body type changes.                   |
| `response.added`                | additive    | info     | A new status code is documented.                     |
| `type.removed`                  | breaking    | major    | A named type disappears.                             |
| `type.kind.changed`             | breaking    | critical | A type's fundamental kind changes.                   |
| `type.enum.removed`             | breaking    | major    | A scalar type drops enum values.                     |
| `type.enum.added`               | behavioral  | minor    | A scalar type gains enum values.                     |
| `type.added`                    | additive    | info     | A new named type appears.                            |
| `field.removed`                 | breaking    | major    | A field disappears from a type.                      |
| `field.added` (required)        | breaking    | major    | A new required field.                                |
| `field.added` (optional)        | additive    | info     | A new optional field.                                |
| `field.type.changed`            | breaking    | major    | A field's type changes.                              |
| `field.arity.changed`           | breaking    | major    | A field flips between scalar and array.              |
| `field.required.added`          | breaking    | major    | An existing field becomes required.                  |
| `field.required.relaxed`        | behavioral  | minor    | An existing field becomes optional.                  |
| `field.nullable.added`          | behavioral  | major    | A field becomes nullable.                            |
| `field.nullable.removed`        | additive    | info     | A field is guaranteed non-null.                      |
| `field.deprecated`              | behavioral  | minor    | A field newly marked deprecated.                     |

### 3.4 Impact attribution rules

- **Field-scoped** changes affect only consumers whose `readsFields` /
  `writesFields` include that `Type.field` path.
- **Type-scoped** changes (no field path) affect consumers that reference any
  field of that type.
- **Parameter-scoped** changes affect consumers that set that parameter —
  except *adding* a parameter, which affects every caller of the endpoint.
- **Endpoint-scoped** changes affect every consumer that calls the endpoint.

---

## 4. Magnitude & verdict

Each impacted (change, consumer) pair contributes `severityWeight ×
criticalityWeight` to a raw score. The raw score is compressed onto a 0–10
Richter-like scale via `10 · (1 − 1/(1 + raw/12))`, so a handful of critical
breaks dominate without a long tail of `info` changes saturating the meter.

| Verdict   | Condition                                    |
|-----------|----------------------------------------------|
| `stable`  | no breaking and no behavioral changes        |
| `tremor`  | minor behavioral drift only                  |
| `shaken`  | magnitude ≥ 3.0 or ≥ 1 breaking change       |
| `rupture` | magnitude ≥ 6.0 or ≥ 3 breaking changes      |

## 5. Exit codes

| Code | Meaning                                             |
|------|-----------------------------------------------------|
| `0`  | stable — safe to ship.                              |
| `1`  | shaken — behavioral changes present.                |
| `2`  | rupture — breaking changes present.                 |
| `3`  | usage / IO error (bad flags, unreadable input).     |

`-fail-on-behavioral` promotes behavioral-only shifts from `1` to `2` for
stricter pipelines.
