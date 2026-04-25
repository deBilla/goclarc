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
| `--db` | `postgres` | Database adapter: `postgres` \| `mongo` \| `rtdb` |
| `--module-path` | `github.com/<user>/<name>` | Go module path written into `go.mod` |
| `--port` | `3001` | Default HTTP port in the generated `Config` struct |

## Examples

```bash
# PostgreSQL project (default)
goclarc new my-api --module-path github.com/acme/my-api

# MongoDB project
goclarc new my-api --db mongo --module-path github.com/acme/my-api

# Firebase RTDB project
goclarc new my-api --db rtdb --module-path github.com/acme/my-api

# Custom port
goclarc new my-api --module-path github.com/acme/my-api --port 8080
```

## Generated Structure

```
my-api/
  cmd/
    api/
      main.go                 # Gin router, DB connect, middleware, graceful shutdown
  internal/
    core/
      config/
        config.go             # Config struct (caarlos0/env), adapter-specific fields
      db/
        db.go                 # Connect() — single shared pool/client for all modules
      errors/
        errors.go             # AppError, sentinel errors, HTTPStatus()
      response/
        response.go           # OK(), Created(), NoContent(), Fail()
    middleware/
      auth.go                 # x-user-id header auth stub
      error.go                # Maps errors.Is(ErrNotFound) → 404, AppError → status
      logger.go               # Zap request logger
  go.mod                      # Includes the correct DB driver for --db
  .gitignore
  .env.example                # Adapter-specific env var examples
```

## Generated `main.go`

The database connection is opened once at startup and the single pool is passed to every module's `NewRepository`:

```go
// PostgreSQL example (--db postgres)
pool, err := db.Connect(context.Background(), cfg.DatabaseURL)
if err != nil {
    logger.Fatal("database", zap.Error(err))
}
defer pool.Close()
logger.Info("database connected")

// TODO: wire your modules here.
//   userRepo    := user.NewRepository(pool)
//   cartRepo    := cart.NewRepository(pool)   // same pool — one connection
//   userService := user.NewService(userRepo)
//   user.RegisterRoutes(v1, userHandler, middleware.Auth())
```

For MongoDB (`--db mongo`):

```go
mongoClient, err := db.Connect(context.Background(), cfg.DatabaseURL)
defer mongoClient.Disconnect(context.Background())

//   userRepo := user.NewRepository(mongoClient.Database("my-api"))
//   cartRepo := cart.NewRepository(mongoClient.Database("my-api"))
```

For Firebase RTDB (`--db rtdb`):

```go
rtdbClient, err := db.Connect(context.Background(), cfg.CredentialsFile, cfg.DatabaseURL)

//   userRepo := user.NewRepository(rtdbClient)
```

## Generated `db.go`

`internal/core/db/db.go` contains a single `Connect()` function tuned for the chosen adapter:

| Adapter | Returns | Connection settings |
|---|---|---|
| `postgres` | `*pgxpool.Pool` | MaxConns 25, MinConns 2, 30m lifetime, ping verified |
| `mongo` | `*mongo.Client` | 10s connect timeout, 5s server selection, ping verified |
| `rtdb` | `*db.Client` | Firebase app init from credentials file |

## Generated Config

Config fields are generated to match the adapter:

**`--db postgres` / `--db mongo`**
```go
type Config struct {
    Port        int    `env:"PORT"         envDefault:"3001"`
    DatabaseURL string `env:"DATABASE_URL" required:"true"`
}
```

**`--db rtdb`**
```go
type Config struct {
    Port            int    `env:"PORT"                  envDefault:"3001"`
    DatabaseURL     string `env:"FIREBASE_DATABASE_URL" required:"true"`
    CredentialsFile string `env:"FIREBASE_CREDENTIALS"  required:"true"`
}
```

## Generated `.env.example`

```bash
# postgres
DATABASE_URL=postgres://user:password@localhost:5432/my-api?sslmode=disable

# mongo
DATABASE_URL=mongodb://localhost:27017/my-api

# rtdb
FIREBASE_DATABASE_URL=https://my-api-default-rtdb.firebaseio.com
FIREBASE_CREDENTIALS=./credentials.json
```
