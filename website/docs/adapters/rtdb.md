---
sidebar_position: 3
---

# Firebase RTDB Adapter

```bash
goclarc module settings --db rtdb --schema schemas/settings.yaml
```

## What Gets Generated

```
internal/modules/settings/repository.go   ← Firebase RTDB implementation
```

## Repository Pattern

The generated repository uses [firebase-admin-go/v4](https://pkg.go.dev/firebase.google.com/go/v4):

```go
type repository struct {
    ref *db.Ref
}

func NewRepository(client *db.Client) Repository {
    return &repository{ref: client.NewRef("settings")}
}

func (r *repository) GetByID(ctx context.Context, id string) (*Entity, error) {
    var data map[string]interface{}
    if err := r.ref.Child(id).Get(ctx, &data); err != nil {
        return nil, fmt.Errorf("settings.repository.GetByID: %w", err)
    }
    if data == nil {
        return nil, fmt.Errorf("settings not found")
    }
    return mapToEntity(data, id), nil
}
```

## Key Differences from Relational Adapters

| Feature | Postgres / MongoDB | Firebase RTDB |
|---|---|---|
| Primary key type | `string` (UUID or ObjectID hex) | RTDB `Push()` key |
| Timestamps | `time.Time` / RFC3339 | `int64` unix milliseconds |
| JSON fields | `json.RawMessage` / `bson.M` | `map[string]interface{}` |
| Create | INSERT / InsertOne | `ref.Push()` |
| List | SELECT * / Find | `ref.Get()` → map iteration |
| Update | UPDATE / FindOneAndUpdate | `ref.Child(id).Update()` |

## Timestamps as Unix Milliseconds

RTDB stores timestamps as `int64` (milliseconds since epoch). There is no `time.Time` conversion — the View type is also `int64`:

```go
// Schema: type: timestamp
// Entity
CreatedAt int64

// View
CreatedAt int64
```

To display as a human-readable date, convert on the client side.

## `mapToEntity` Conversion

RTDB returns `map[string]interface{}` from reads. The generated converter type-asserts each field:

```go
func mapToEntity(data map[string]interface{}, id string) *Entity {
    e := &Entity{}
    e.ID = id
    if v, ok := data["email"]; ok && v != nil {
        if cast, ok := v.(string); ok {
            e.Email = cast
        }
    }
    // ...
    return e
}
```

## Required Imports

```go
import "firebase.google.com/go/v4/db"
```

```bash
go get firebase.google.com/go/v4
```

## Connection Setup

```go
app, err := firebase.NewApp(ctx, nil)
if err != nil {
    logger.Fatal("firebase init", zap.Error(err))
}
dbClient, err := app.DatabaseWithURL(ctx, cfg.FirebaseDBURL)
if err != nil {
    logger.Fatal("firebase db", zap.Error(err))
}

settingsRepo    := settings.NewRepository(dbClient)
settingsService := settings.NewService(settingsRepo)
settingsHandler := settings.NewHandler(settingsService)
settings.RegisterRoutes(v1, settingsHandler, middleware.Auth())
```

Firebase authentication is handled via Application Default Credentials (ADC) — set `GOOGLE_APPLICATION_CREDENTIALS` in your environment.
