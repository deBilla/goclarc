# goclarc

[![Go Version](https://img.shields.io/badge/go-1.22+-00ADD8?logo=go)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![CI](https://github.com/deBilla/goclarc/actions/workflows/ci.yml/badge.svg)](https://github.com/deBilla/goclarc/actions/workflows/ci.yml)
[![Docs](https://img.shields.io/badge/docs-GitHub%20Pages-brightgreen)](https://deBilla.github.io/goclarc/)

> **NestJS-style developer experience for Go.** Scaffold production-ready Clean Architecture APIs in seconds.

`goclarc` is a CLI that generates fully typed, database-connected Go modules from a simple YAML schema. Define your domain once — get handler, service, repository, entity, DTO, and routes, wired and ready to extend.

---

## Install

```bash
go install github.com/deBilla/goclarc/cmd/goclarc@latest
```

Verify:

```bash
goclarc --version
```

---

## Quick Start

```bash
# 1. Scaffold a new project
goclarc new my-api --module-path github.com/you/my-api

cd my-api

# 2. Define your domain in YAML
cat > schemas/user.yaml << 'EOF'
module: user
table: users
fields:
  - name: id
    type: uuid
    primary: true
    auto: true
  - name: email
    type: string
    required: true
  - name: name
    type: string
    required: true
  - name: created_at
    type: timestamp
    auto: true
EOF

# 3. Generate the module
goclarc module user --db postgres --schema schemas/user.yaml

# 4. Install dependencies and run
go mod tidy
go run ./cmd/api
```

---

## What Gets Generated

`goclarc module` produces 6 files per module (+ SQL for Postgres):

| File | Purpose |
|---|---|
| `entity.go` | Domain model (`Entity`) and API view (`View`) with `ToView()` |
| `dto.go` | `CreateRequest`, `UpdateRequest` with Gin binding tags |
| `repository.go` | `Repository` interface + DB-specific implementation |
| `service.go` | `Service` interface + business logic stub |
| `handler.go` | Gin CRUD handlers (Create, GetByID, List, Update, Delete) |
| `routes.go` | `RegisterRoutes()` wiring all 5 endpoints |
| `*s.sql` *(Postgres)* | sqlc-annotated CRUD queries |

`goclarc new` produces a full project skeleton:

```
my-api/
  cmd/api/main.go                   # Gin server, graceful shutdown
  internal/core/config/config.go    # env-based Config (caarlos0/env)
  internal/core/errors/errors.go    # AppError, sentinel errors
  internal/core/response/           # OK(), Created(), Fail() helpers
  internal/middleware/              # Logger (zap), ErrorHandler, Auth stub
  go.mod / .gitignore / .env.example
```

---

## Database Adapters

| Flag | Database | Driver | Notes |
|---|---|---|---|
| `--db postgres` | PostgreSQL | pgx/v5 + sqlc | Generates `.sql` query file |
| `--db mongo` | MongoDB | mongo-driver/v2 | Uses `bson.M` for JSON fields |
| `--db rtdb` | Firebase RTDB | firebase-admin/v4 | Timestamps stored as unix ms |

One project can mix adapters — each module targets exactly one database.

---

## Schema Format

```yaml
module: product           # package name (snake_case)
table: products           # DB table / collection / RTDB ref

fields:
  - name: id
    type: uuid
    primary: true
    auto: true            # skipped from Create/Update params

  - name: title
    type: string
    required: true        # adds binding:"required" to CreateRequest

  - name: price
    type: float
    required: true

  - name: stock
    type: int
    nullable: true        # *int32 in Entity, nil-safe in ToView()

  - name: metadata
    type: json            # json.RawMessage / bson.M / map[string]interface{}
    nullable: true

  - name: created_at
    type: timestamp
    auto: true
```

**Supported types:** `uuid` · `string` · `int` · `int64` · `float` · `bool` · `timestamp` · `json` · `string[]`

---

## CLI Reference

### `goclarc new [project-name]`

| Flag | Default | Description |
|---|---|---|
| `--module-path` | `github.com/<user>/<name>` | Go module path |
| `--port` | `3001` | Default HTTP port in generated config |

### `goclarc module [name]`

| Flag | Short | Default | Description |
|---|---|---|---|
| `--schema` | `-s` | *(required)* | Path to schema YAML |
| `--db` | | `postgres` | `postgres` \| `mongo` \| `rtdb` |
| `--out-dir` | `-o` | `internal/modules/<name>` | Output directory |
| `--query-dir` | | `schemas/queries` | SQL output dir (postgres only) |
| `--force` | `-f` | `false` | Overwrite existing files |
| `--dry-run` | | `false` | Print output without writing |
| `--module-path` | | *(from go.mod)* | Go module path |

---

## Architecture

Generated code follows strict Clean Architecture layering:

```
HTTP Request
     │
     ▼
  Handler          (parse request, call service, write response)
     │
     ▼
  Service          (business logic, orchestrates repositories)
     │
     ▼
  Repository       (data access — DB calls only)
     │
     ▼
  Database
```

- Handlers depend on **Service interfaces**, never concrete types
- Services depend on **Repository interfaces**, never concrete types
- Errors flow up via `c.Error(err)` to a global error middleware
- All I/O functions take `ctx context.Context` as first argument

---

## Documentation

Full documentation at **[deBilla.github.io/goclarc](https://deBilla.github.io/goclarc/)**

- [Getting Started](https://deBilla.github.io/goclarc/docs/getting-started)
- [Schema Reference](https://deBilla.github.io/goclarc/docs/schema/overview)
- [CLI Reference](https://deBilla.github.io/goclarc/docs/commands/new)
- [Database Adapters](https://deBilla.github.io/goclarc/docs/adapters/postgres)
- [Examples](https://deBilla.github.io/goclarc/docs/examples)

---

## Contributing

Contributions are welcome. See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

---

## License

MIT — see [LICENSE](LICENSE).
