---
sidebar_position: 2
---

# Field Types

goclarc supports 9 schema types. Each maps to different Go types depending on the database adapter and nullable flag.

## Type Reference

### `uuid`

A UUID string identifier. Typically used for primary keys.

| | PostgreSQL | MongoDB | Firebase RTDB |
|---|---|---|---|
| **Entity** | `string` | `string` | `string` |
| **View** | `string` | `string` | `string` |
| **Create** | `string` | `string` | `string` |
| **Update** | `*string` | `*string` | `*string` |

---

### `string`

A text field.

| | PostgreSQL (non-null) | PostgreSQL (nullable) | MongoDB | RTDB |
|---|---|---|---|---|
| **Entity** | `string` | `*string` | `string` / `*string` | `string` / `*string` |
| **View** | `string` | `string` | `string` | `string` |
| **Create** | `string` | `*string` | `string` / `*string` | `string` / `*string` |
| **Update** | `*string` | `*string` | `*string` | `*string` |

---

### `int`

A 32-bit integer.

| | Non-nullable | Nullable |
|---|---|---|
| **Entity** | `int32` | `*int32` |
| **View** | `int32` | `int32` |
| **Create** | `int32` | `*int32` |
| **Update** | `*int32` | `*int32` |

> **Note:** Firebase RTDB uses plain `int` (not `int32`).

---

### `int64`

A 64-bit integer.

| | Non-nullable | Nullable |
|---|---|---|
| **Entity** | `int64` | `*int64` |
| **View** | `int64` | `int64` |
| **Create** | `int64` | `*int64` |
| **Update** | `*int64` | `*int64` |

---

### `float`

A 64-bit floating-point number.

| | Non-nullable | Nullable |
|---|---|---|
| **Entity** | `float64` | `*float64` |
| **View** | `float64` | `float64` |
| **Create** | `float64` | `*float64` |
| **Update** | `*float64` | `*float64` |

---

### `bool`

A boolean value.

| | Non-nullable | Nullable |
|---|---|---|
| **Entity** | `bool` | `*bool` |
| **View** | `bool` | `bool` |
| **Create** | `bool` | `*bool` |
| **Update** | `*bool` | `*bool` |

---

### `timestamp`

A date-time value. View is always a string (RFC3339).

| | PostgreSQL | MongoDB | Firebase RTDB |
|---|---|---|---|
| **Entity (non-null)** | `time.Time` | `time.Time` | `int64` (unix ms) |
| **Entity (nullable)** | `*time.Time` | `*time.Time` | `int64` |
| **View** | `string` (RFC3339) | `string` (RFC3339) | `int64` |
| **Create** | `time.Time` | `time.Time` | `int64` |
| **Update** | `*time.Time` | `*time.Time` | `*int64` |

Typically used with `auto: true` for `created_at` / `updated_at` fields.

---

### `json`

Arbitrary JSON / document data.

| | PostgreSQL | MongoDB | Firebase RTDB |
|---|---|---|---|
| **Entity** | `json.RawMessage` | `bson.M` | `map[string]interface{}` |
| **View** | `json.RawMessage` | `bson.M` | `map[string]interface{}` |
| **Create** | `json.RawMessage` | `bson.M` | `map[string]interface{}` |
| **Update** | `json.RawMessage` | `bson.M` | `map[string]interface{}` |

---

### `string[]`

An array of strings.

| | All adapters |
|---|---|
| **Entity** | `[]string` |
| **View** | `[]string` |
| **Create** | `[]string` |
| **Update** | `[]string` |

---

## Update Params

For non-auto fields, `UpdateParams` always uses pointer types to allow partial updates — only non-nil values are applied in the repository:

```go
type UpdateParams struct {
    ID    string
    Email *string   // nil = don't update this field
    Name  *string
    Age   *int32
}
```

The generated repository applies a `COALESCE(@field, field)` pattern for Postgres, and checks `if p.Field != nil` for MongoDB and RTDB.
