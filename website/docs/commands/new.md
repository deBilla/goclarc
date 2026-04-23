---
sidebar_position: 1
---

# goclarc new

Scaffold a new Clean Architecture Go project skeleton.

## Usage

```bash
goclarc new [project-name] [flags]
```

## Arguments

| Argument | Description |
|---|---|
| `project-name` | Name of the project directory to create (required) |

## Flags

| Flag | Default | Description |
|---|---|---|
| `--module-path` | `github.com/<user>/<name>` | Go module path written into `go.mod` |
| `--port` | `3001` | Default HTTP port in the generated `Config` struct |

## Examples

```bash
# Minimal — uses a placeholder module path
goclarc new my-api

# With explicit module path and port
goclarc new my-api \
  --module-path github.com/acme/my-api \
  --port 8080
```

## Generated Structure

```
my-api/
  cmd/
    api/
      main.go                 # Gin router, middleware wiring, graceful shutdown
  internal/
    core/
      config/
        config.go             # Config struct (caarlos0/env)
      errors/
        errors.go             # AppError, sentinel errors, HTTPStatus()
      response/
        response.go           # OK(), Created(), NoContent(), Fail()
    middleware/
      auth.go                 # x-user-id header auth stub
      error.go                # Maps AppError → HTTP status + JSON envelope
      logger.go               # Zap request logger
  go.mod
  .gitignore
  .env.example
```

## Generated `main.go`

```go
func main() {
    // ...
    v1 := r.Group("/api/v1")

    // TODO: wire your modules here.
    // Example:
    //   userRepo    := user.NewRepository(pool)
    //   userService := user.NewService(userRepo)
    //   userHandler := user.NewHandler(userService)
    //   user.RegisterRoutes(v1, userHandler, middleware.Auth())
}
```

After running `goclarc new`, add your DB connection and wire each generated module at the `TODO` point.

## Generated Config

The config struct uses `caarlos0/env` for environment-based loading:

```go
type Config struct {
    Port        int    `env:"PORT"          envDefault:"3001"`
    DatabaseURL string `env:"DATABASE_URL"  required:"true"`
    RedisHost   string `env:"REDIS_HOST"    envDefault:"127.0.0.1"`
    RedisPort   int    `env:"REDIS_PORT"    envDefault:"6379"`
}
```

Extend this struct with any secrets or config values your project needs.
