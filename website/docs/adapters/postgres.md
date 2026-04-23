---
sidebar_position: 1
---

# PostgreSQL Adapter

```bash
goclarc module user --db postgres --schema schemas/user.yaml
```

## What Gets Generated

In addition to the standard 6 module files, the Postgres adapter generates two extra files:

```
internal/modules/user/repository.go   ← pgxpool raw-query implementation
schemas/queries/users.sql             ← CRUD SQL for reference / future use
db/migrations/001_create_users.sql    ← CREATE TABLE migration (auto-generated)
```

## Repository Pattern

The generated repository uses [pgx/v5](https://github.com/jackc/pgx) directly — no sqlc or ORM dependency. All CRUD queries are embedded as raw SQL strings.

```go
type repository struct {
    pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
    return &repository{pool: pool}
}

func (r *repository) GetByID(ctx context.Context, id string) (*Entity, error) {
    row := r.pool.QueryRow(ctx,
        `SELECT id, email, name, created_at FROM users WHERE id = $1`,
        id,
    )
    entity, err := scanEntity(row)
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, fmt.Errorf("user not found")
        }
        return nil, fmt.Errorf("user.repository.GetByID: %w", err)
    }
    return entity, nil
}
```

pgx/v5 natively scans UUIDs to `string`, `TIMESTAMPTZ` to `time.Time`, and `NULL` columns to pointer types (`*string`, `*int32`, etc.) — no intermediate helper types required.

## Migration File

goclarc automatically generates a `CREATE TABLE` migration alongside your module:

```sql title="db/migrations/001_create_users.sql"
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL,
    name TEXT NOT NULL,
    age INT,
    is_active BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);
```

The migration directory defaults to `db/migrations/`. Override it with `--migration-dir`:

```bash
goclarc module user --db postgres --schema schemas/user.yaml \
  --migration-dir supabase/migrations
```

## Required Imports

The generated repository only needs:

```go
import (
    "context"
    "errors"
    "fmt"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)
```

Add the pgx dependency:

```bash
go get github.com/jackc/pgx/v5
```

No sqlc installation or code generation step is required.

## Update Pattern

Updates use `COALESCE` so only provided fields are changed — `nil` pointer fields leave the DB column unchanged:

```sql
UPDATE users SET
  email = COALESCE($2, email),
  name  = COALESCE($3, name)
WHERE id = $1
RETURNING id, email, name, created_at
```

## Connection Setup

In `cmd/api/main.go`, connect to Postgres before wiring your repository:

```go
pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
if err != nil {
    logger.Fatal("postgres", zap.Error(err))
}
defer pool.Close()

userRepo    := user.NewRepository(pool)
userService := user.NewService(userRepo)
userHandler := user.NewHandler(userService)
user.RegisterRoutes(v1, userHandler, middleware.Auth())
```
