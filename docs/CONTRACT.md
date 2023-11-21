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
