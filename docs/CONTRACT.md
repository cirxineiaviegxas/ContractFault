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
