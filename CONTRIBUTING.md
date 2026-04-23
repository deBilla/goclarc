# Contributing to goclarc

Thank you for your interest in contributing! This guide will help you get started.

---

## Development Setup

```bash
git clone https://github.com/deBilla/goclarc.git
cd goclarc
go mod download
go build ./cmd/goclarc
go test ./...
```

---

## Project Structure

```
cmd/goclarc/          # Binary entry point
internal/
  assets/             # Embedded templates (//go:embed)
    templates/
      module/         # Per-module code templates
      project/        # New project skeleton templates
  cmd/                # Cobra command implementations
  gen/                # Template context + rendering engine
  schema/             # YAML parser + DB adapter type mappings
```

---

## How to Add a New Database Adapter

1. Add a new adapter struct in `internal/schema/types.go` implementing `Adapter`:
   ```go
   type myAdapter struct{}
   func (a *myAdapter) Name() string { return "mydb" }
   func (a *myAdapter) DBImports(modulePath string) []string { ... }
   func (a *myAdapter) Map(fieldType string, nullable bool) GoTypeInfo { ... }
   ```

2. Register it in `AdapterFor()` in `types.go`.

3. Add a new template `internal/assets/templates/module/repository.mydb.go.tmpl`.

4. Reference the new `--db mydb` option in `internal/cmd/module.go`.

5. Add tests covering your type mappings.

---

## How to Modify Templates

Templates live in `internal/assets/templates/`. They use Go's `text/template` syntax.

Available template context fields are documented in `internal/gen/context.go`.

After editing templates, rebuild and run a dry-run to verify:

```bash
go build -o /tmp/goclarc ./cmd/goclarc
/tmp/goclarc module test --db postgres --schema path/to/schema.yaml --dry-run
```

---

## Running Tests

```bash
# All tests
go test ./...

# With race detector
go test -race ./...

# Specific package
go test ./internal/gen/...
```

---

## Pull Request Guidelines

- Open an issue first for anything larger than a bug fix or typo
- Keep PRs focused — one feature or fix per PR
- Add or update tests for any changed behaviour
- Run `go vet ./...` and `go test ./...` before submitting
- Update docs in `website/docs/` if you change CLI flags or behaviour

---

## Reporting Issues

Use [GitHub Issues](https://github.com/deBilla/goclarc/issues). Include:
- goclarc version (`goclarc --version`)
- Go version (`go version`)
- Schema YAML (if relevant)
- Full error output
