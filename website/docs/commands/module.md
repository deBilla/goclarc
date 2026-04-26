---
sidebar_position: 2
---

# goclarc module

Generate a typed Clean Architecture module from a YAML schema.

## Usage

```bash
goclarc module [name] [flags]
```

## Arguments

| Argument | Description |
|---|---|
| `name` | Module name — used as the Go package name and default output directory |

## Flags

| Flag | Short | Default | Description |
|---|---|---|---|
| `--schema` | `-s` | *(required)* | Path to the schema YAML file |
| `--db` | | `postgres` | Database adapter: `postgres` \| `mongo` \| `rtdb` |
| `--out-dir` | `-o` | `internal/modules/<name>` | Output directory for generated Go files |
| `--query-dir` | | `schemas/queries` | Directory for the generated `.sql` file (postgres only) |
| `--migration-dir` | | `db/migrations` | Directory for the generated `CREATE TABLE` migration (postgres only) |
| `--force` | `-f` | `false` | Overwrite existing files without prompting |
| `--dry-run` | | `false` | Print generated output to stdout — write nothing |
| `--reset` | | `false` | Delete all files previously generated for this module |
| `--module-path` | | *(from go.mod)* | Go module path for import statements |
| `--swagger` | | `false` | Generate `docs/<name>.openapi.yaml` OpenAPI 3.0 spec for this module |
| `--mode` | | *(full stack)* | `repo` — generate entity, dto, repository, and migration only (no service, handler, or routes) |
| `--parent` | | *(none)* | Parent route prefix for nested resources, e.g. `/workspaces/:workspaceId`. The module routes are mounted under this prefix. |

## Examples

```bash
# PostgreSQL module
goclarc module user --db postgres --schema schemas/user.yaml

# MongoDB module
goclarc module post --db mongo --schema schemas/post.yaml

# Firebase RTDB module
goclarc module settings --db rtdb --schema schemas/settings.yaml

# Preview without writing
goclarc module user --db postgres --schema schemas/user.yaml --dry-run

# Custom output and migration directories
goclarc module user --db postgres --schema schemas/user.yaml \
  --out-dir pkg/modules/user \
  --migration-dir supabase/migrations

# Overwrite existing files after schema changes
goclarc module user --db postgres --schema schemas/user.yaml --force

# Remove all generated files for a module
goclarc module user --db postgres --schema schemas/user.yaml --reset

# Generate OpenAPI 3.0 spec alongside the module
goclarc module user --db postgres --schema schemas/user.yaml --swagger
```

## Generated Files

For `goclarc module user --db postgres`:

```
internal/modules/user/
  entity.go       ← Entity struct, View struct, ToView()
  dto.go          ← CreateRequest, UpdateRequest, CreateParams, UpdateParams
  repository.go   ← Repository interface + pgxpool raw-query implementation
  service.go      ← Service interface + implementation
  handler.go      ← Gin CRUD handlers (Create, GetByID, List, Update, Delete)
  routes.go       ← RegisterRoutes() wiring all 5 endpoints
schemas/queries/
  users.sql       ← CRUD SQL reference (postgres only)
db/migrations/
  001_create_users.sql  ← CREATE TABLE migration (postgres only)
```

With `--swagger`, one additional file is generated:

```
docs/
  user.openapi.yaml   ← Complete OpenAPI 3.0 spec for the user module
```

Merge it into `docs/openapi.yaml` (created by `goclarc new --swagger`) to update the live spec:

```bash
cp docs/user.openapi.yaml docs/openapi.yaml
# or merge multiple modules with yq:
# yq '. *+ load("docs/user.openapi.yaml")' docs/openapi.yaml > docs/openapi.yaml
```

## Resetting a Module

The `--reset` flag deletes every file that would have been generated for the module and removes the module directory if it becomes empty. Useful when you want to start fresh after renaming fields or changing the adapter:

```bash
# Remove all generated files, then regenerate with updated schema
goclarc module user --db postgres --schema schemas/user.yaml --reset
goclarc module user --db postgres --schema schemas/user.yaml
```

## Repository-Only Mode

Use `--mode repo` when a module needs a data layer but no HTTP surface — join tables, audit logs, or internal storage modules:

```bash
goclarc module workspace_member --db postgres --schema schemas/workspace_member.yaml --mode repo
```

Generated files (postgres):

```
internal/modules/workspace_member/
  entity.go        ← Entity, View, ToView()
  dto.go           ← CreateParams, UpdateParams
  repository.go    ← Repository interface + implementation
db/migrations/
  001_create_workspace_members.sql
```

`service.go`, `handler.go`, and `routes.go` are **not** generated.

## Nested Resources

Use `--parent` to mount a module's routes under a parent resource path:

```bash
goclarc module sheet --db postgres --schema schemas/sheet.yaml \
  --parent "/workspaces/:workspaceId"
```

The generated `RegisterRoutes` becomes:

```go
func RegisterRoutes(rg *gin.RouterGroup, h *Handler, middlewares ...gin.HandlerFunc) {
    g := rg.Group("/workspaces/:workspaceId/sheets", middlewares...)
    g.POST("", h.Create)
    g.GET("", h.List)
    g.GET("/:id", h.GetByID)
    g.PATCH("/:id", h.Update)
    g.DELETE("/:id", h.Delete)
}
```

The parent param name must not be `:id` — that name is reserved for the module's own resource identifier. goclarc rejects conflicting names at generation time:

```
Error: --parent "/foo/:id" contains :id which conflicts with the module's own /:id param
```

:::caution
`--reset` permanently deletes files. Any custom business logic you added to `service.go` or extra queries in `repository.go` will be lost.
:::

## Module Name vs Schema Module Field

The `name` argument to `goclarc module` is used as the Go package name and determines the output directory. The `module` field in the schema YAML sets the **table/collection name** default and is overridden by the command argument.

This means you can reuse a schema with a different module name:

```bash
# Generates package "admin_user" using the same user schema
goclarc module admin_user --db postgres --schema schemas/user.yaml
```

## Registered Routes

Every generated module exposes these 5 endpoints:

| Method | Path | Handler |
|---|---|---|
| `POST` | `/<plural>` | `Create` |
| `GET` | `/<plural>` | `List` |
| `GET` | `/<plural>/:id` | `GetByID` |
| `PATCH` | `/<plural>/:id` | `Update` |
| `DELETE` | `/<plural>/:id` | `Delete` |

Where `<plural>` is the pluralised kebab-case of the module name (e.g., `user` → `users`, `blog-post` → `blog-posts`).

## Module Path Detection

If `--module-path` is omitted, goclarc reads the `module` directive from `go.mod` in the current directory. Run `goclarc module` from the project root for automatic detection.
