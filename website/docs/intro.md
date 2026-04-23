---
sidebar_position: 1
slug: /
---

# What is goclarc?

**goclarc** is a CLI tool that scaffolds production-ready Go APIs following Clean Architecture — similar to what `@nestjs/cli` does for TypeScript, but for Go.

## The Problem

Starting a new Go API means writing the same boilerplate every time:
- Entity structs and view converters
- DTO structs with Gin binding tags
- Repository interfaces and implementations for your DB of choice
- Service interfaces and method stubs
- Gin handler functions
- Route registration

Do this across 10 modules and you've spent a day writing code that adds no value.

## The Solution

Define your domain in a YAML schema file:

```yaml
module: user
fields:
  - name: id
    type: uuid
    primary: true
    auto: true
  - name: email
    type: string
    required: true
```

Run one command:

```bash
goclarc module user --db postgres --schema schemas/user.yaml
```

Get six typed Go files, all wired together:

```
internal/modules/user/
  entity.go       ← Entity + View structs, ToView()
  dto.go          ← CreateRequest, UpdateRequest with binding tags
  repository.go   ← Repository interface + pgx implementation
  service.go      ← Service interface + business logic stub
  handler.go      ← Gin CRUD handlers
  routes.go       ← RegisterRoutes()
schemas/queries/
  users.sql       ← sqlc-annotated CRUD queries (postgres only)
```

## Architecture

Generated code strictly follows Clean Architecture:

```
HTTP Request
     │
     ▼
  Handler          parse request, delegate to service, write response
     │  (interface)
     ▼
  Service          business logic, orchestrate repositories
     │  (interface)
     ▼
  Repository       data access — one database, one job
     │
     ▼
  Database
```

**Rules enforced in every generated file:**
- Handlers depend on Service **interfaces** — never concrete types
- Services depend on Repository **interfaces** — never concrete types
- All I/O methods take `ctx context.Context` as the first parameter
- Errors flow through `c.Error(err)` to a global middleware — never inline `c.JSON` for errors

## What's Included

| Command | What it does |
|---|---|
| `goclarc new` | Scaffold a new Go project skeleton (Gin, config, middleware, errors) |
| `goclarc module` | Generate a typed module from a YAML schema for Postgres, MongoDB, or RTDB |

## Next Steps

- [Install goclarc](/docs/installation)
- [Build your first API in 5 minutes](/docs/getting-started)
- [Schema reference](/docs/schema/overview)
