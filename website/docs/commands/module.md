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
| `--force` | `-f` | `false` | Overwrite existing files without prompting |
| `--dry-run` | | `false` | Print generated output to stdout — write nothing |
| `--module-path` | | *(from go.mod)* | Go module path for import statements |

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

# Custom output directory
goclarc module user --db postgres --schema schemas/user.yaml \
  --out-dir pkg/modules/user

# Overwrite existing files
goclarc module user --db postgres --schema schemas/user.yaml --force
```

## Generated Files

For `goclarc module user --db postgres`:

```
internal/modules/user/
  entity.go       ← Entity struct, View struct, ToView()
  dto.go          ← CreateRequest, UpdateRequest, CreateParams, UpdateParams
  repository.go   ← Repository interface + pgx implementation
  service.go      ← Service interface + implementation
  handler.go      ← Gin CRUD handlers (Create, GetByID, List, Update, Delete)
  routes.go       ← RegisterRoutes() wiring all 5 endpoints
schemas/queries/
  users.sql       ← sqlc CRUD query file (postgres only)
```

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
