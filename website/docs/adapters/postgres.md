---
sidebar_position: 1
---

# PostgreSQL Adapter

```bash
goclarc module user --db postgres --schema schemas/user.yaml
```

## What Gets Generated

In addition to the standard 6 module files, the Postgres adapter generates a SQL query file:

```
internal/modules/user/repository.go   ← pgxpool + sqlcdb implementation
schemas/queries/users.sql             ← sqlc-annotated CRUD queries
```

## Repository Pattern

The generated repository uses [sqlc](https://sqlc.dev/) for type-safe queries and [pgx/v5](https://github.com/jackc/pgx) for the database driver:

```go
type repository struct {
    q *sqlcdb.Queries
}

func NewRepository(pool *pgxpool.Pool) Repository {
    return &repository{q: sqlcdb.New(pool)}
}

func (r *repository) GetByID(ctx context.Context, id string) (*Entity, error) {
    row, err := r.q.GetUserByID(ctx, id)
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, fmt.Errorf("user not found")
        }
        return nil, fmt.Errorf("user.repository.GetByID: %w", err)
    }
    return rowToEntity(row), nil
}
```

## Setting Up sqlc

goclarc writes the SQL query file. You still need to run sqlc to generate the Go database layer.

**1. Install sqlc:**

```bash
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

**2. Create `sqlc.yaml` in your project root:**

```yaml
version: "2"
sql:
  - engine: postgresql
    queries: schemas/queries/
    schema: db/migrations/
    gen:
      go:
        package: sqlcdb
        out: internal/db/sqlcdb
        sql_package: pgx/v5
```

**3. Add your schema migrations to `db/migrations/`** (the table DDL that matches your YAML fields).

**4. Generate:**

```bash
sqlc generate
```

This creates `internal/db/sqlcdb/` with the `Queries` struct that the generated repository uses.

## Generated SQL File

For a `user` module with `email`, `name`, `age`, `is_active` fields, the generated `schemas/queries/users.sql` looks like:

```sql
-- name: CreateUser :one
INSERT INTO users (
  email,
  name,
  age,
  is_active
) VALUES (
  @email,
  @name,
  @age,
  @is_active
)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = @id
LIMIT 1;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY created_at DESC;

-- name: UpdateUser :one
UPDATE users
SET
  email = COALESCE(@email, email),
  name = COALESCE(@name, name),
  age = COALESCE(@age, age),
  is_active = COALESCE(@is_active, is_active)
WHERE id = @id
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = @id;
```

## Required Imports

The generated repository imports:

```go
import (
    "context"
    "errors"
    "fmt"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
    "your-module/internal/db/sqlcdb"
)
```

Add these dependencies:

```bash
go get github.com/jackc/pgx/v5
go get github.com/sqlc-dev/sqlc/cmd/sqlc
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
