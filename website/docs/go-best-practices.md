---
sidebar_position: 9
---

# Go Best Practices

Every file goclarc generates follows the patterns described in [Effective Go](https://go.dev/doc/effective_go) and the conventions established by the Go team. This page explains **why** each pattern is used and **how to apply it** when you extend the generated code — so that your custom additions stay idiomatic.

This is your reference for writing Go the goclarc way.

---

## Naming

### Package names

Package names are lowercase, single-word, and never abbreviated to the point of obscurity. No underscores, no MixedCaps.

```go
// Generated package names
package user
package product
package middleware
package errors
```

When you create shared utilities, follow the same rule:

```go
// Good
package pagination
package validate

// Bad — underscores and mixed case are not Go style
package user_service
package UserService
```

### MixedCaps — the one rule for multi-word names

Go uses `MixedCaps` (or `mixedCaps` for unexported) everywhere. Underscores appear only in test function names and generated SQL.

```go
// Exported: PascalCase
type CreateRequest struct { ... }
func NewService(repo Repository) Service { ... }

// Unexported: camelCase
type repository struct { ... }
func scanEntity(s scanner) (*Entity, error) { ... }
```

### Initialisms stay all-caps

Go treats well-known acronyms as atomic units. goclarc's name converter handles this automatically from your schema field names.

```go
// Correct — initialism preserved
type Entity struct {
    UserID   string    // not UserId
    APIKey   string    // not ApiKey
    HTTPCode int       // not HttpCode
}

// What the name converter does for you:
// "user_id"  → UserID
// "api_key"  → APIKey
// "http_url" → HTTPURL
```

The full initialism set (ACL, API, CPU, DNS, EOF, GUID, HTML, HTTP, HTTPS, ID, IP, JSON, QPS, RAM, RPC, SQL, SSH, TCP, TLS, TTL, UDP, UI, UID, UUID, URI, URL, UTF8, VM, XML) is built into the generator. Custom fields that contain these substrings get the correct casing automatically.

### Getters — drop the `Get`

If a method returns a field value, name it after the field. The `Get` prefix is not idiomatic Go.

```go
// Good
func (u *User) Email() string { return u.email }

// Bad — redundant "Get" prefix
func (u *User) GetEmail() string { return u.email }
```

Setters can use `Set` because there is no ambiguity:

```go
func (u *User) SetEmail(email string) { u.email = email }
```

### Interface names — the "-er" suffix

A single-method interface takes the method name plus `-er`. This is why the internal scan abstraction in the postgres repository is called `scanner` (from `Scan`).

```go
// From the generated postgres repository — a private single-method interface
type scanner interface {
    Scan(dest ...any) error
}

// Standard library examples of the same pattern
// io.Reader   → Read()
// io.Writer   → Write()
// fmt.Stringer → String()
```

When you define your own interfaces, follow this:

```go
// Good
type Validator interface {
    Validate() error
}

type Notifier interface {
    Notify(ctx context.Context, msg string) error
}
```

Multi-method interfaces describe a role rather than an action:

```go
type Repository interface {
    Create(ctx context.Context, p CreateParams) (*Entity, error)
    GetByID(ctx context.Context, id string) (*Entity, error)
    List(ctx context.Context) ([]*Entity, error)
    Update(ctx context.Context, p UpdateParams) (*Entity, error)
    Delete(ctx context.Context, id string) error
}
```

---

## Commentary

### Document every exported identifier

A doc comment sits directly above the declaration with no blank line between them. It starts with the identifier name.

```go
// Service defines the business operations for the User domain.
type Service interface { ... }

// NewService constructs a Service backed by the given Repository.
func NewService(repo Repository) Service { ... }

// ErrNotFound is returned when a requested resource does not exist.
var ErrNotFound = errors.New("not found")
```

### Package comments

Each package has exactly one package-level comment, placed in any file in the package (conventionally the file with the most important content).

```go
// Package errors defines sentinel errors and the AppError type used
// throughout the application for structured error handling.
package errors
```

### What NOT to comment

Skip comments that repeat what the code already says:

```go
// Bad — states the obvious
// scanEntity scans a row into an Entity.
func scanEntity(s scanner) (*Entity, error) { ... }

// Good — explains the non-obvious: why scanner covers both pgx.Row and pgx.Rows
// scanner is satisfied by both pgx.Row and pgx.Rows, allowing scanEntity
// to be called from both QueryRow and the rows.Next() loop without duplication.
type scanner interface {
    Scan(dest ...any) error
}
```

---

## Error Handling

This is the most important section. Effective Go's error model is central to how goclarc works.

### Errors are values — return them as the last result

Every function that can fail returns an `error` as its final return value. Never use panics for expected failure modes.

```go
func (s *service) Create(ctx context.Context, req CreateRequest) (*Entity, error) {
    entity, err := s.repo.Create(ctx, req.ToCreateParams())
    if err != nil {
        return nil, fmt.Errorf("user.service.Create: %w", err)
    }
    return entity, nil
}
```

### Wrap errors with context at every layer

Use `fmt.Errorf("context: %w", err)` when propagating errors. The `%w` verb wraps the original error, preserving the chain for `errors.Is` and `errors.As`.

The wrapping pattern used throughout goclarc is `<module>.<layer>.<Method>`:

```go
// Repository layer
return nil, fmt.Errorf("user.repository.GetByID: %w", err)

// Service layer (wraps the repository error)
return nil, fmt.Errorf("user.service.GetByID: %w", err)
```

This produces readable stack traces in logs:

```
user.service.GetByID: user.repository.GetByID: ERROR: relation "users" does not exist
```

### Sentinel errors and `errors.Is`

Sentinel errors are package-level variables that represent specific failure conditions. Use `errors.Is` to check for them — never compare error strings.

```go
// Defined once in internal/core/errors/errors.go
var (
    ErrNotFound      = errors.New("not found")
    ErrForbidden     = errors.New("forbidden")
    ErrConflict      = errors.New("conflict")
    ErrLimitExceeded = errors.New("limit exceeded")
)
```

The repository wraps sentinels so the full error chain is preserved:

```go
// Repository wraps the sentinel — the chain is intact
if errors.Is(err, pgx.ErrNoRows) {
    return nil, fmt.Errorf("user.repository.GetByID: %w", apperr.ErrNotFound)
}
```

The error middleware then unwraps and checks with `errors.Is`:

```go
// Middleware — errors.Is traverses the whole chain automatically
if errors.Is(err, apperr.ErrNotFound) {
    c.JSON(http.StatusNotFound, ...)
    return
}
```

**Why this matters:** String matching like `strings.Contains(msg, "not found")` breaks the moment a message changes. `errors.Is` is refactor-safe, explicit, and the idiomatic Go way.

### `AppError` — typed errors with user messages

When you need to return both a machine-readable sentinel and a human-readable message, use `AppError`:

```go
// AppError wraps a sentinel with a user-facing message
type AppError struct {
    Cause   error
    Message string
}

func (e *AppError) Error() string { return e.Message }
func (e *AppError) Unwrap() error { return e.Cause }
```

Usage in your service layer:

```go
import apperr "github.com/you/my-api/internal/core/errors"

func (s *service) Create(ctx context.Context, req CreateRequest) (*Entity, error) {
    existing, err := s.repo.GetByEmail(ctx, req.Email)
    if err != nil && !errors.Is(err, apperr.ErrNotFound) {
        return nil, fmt.Errorf("user.service.Create: %w", err)
    }
    if existing != nil {
        return nil, apperr.New(apperr.ErrConflict, "a user with this email already exists")
    }
    ...
}
```

The error middleware detects `AppError` with `errors.As` and formats the response automatically.

### Never silently discard errors

The blank identifier `_` is acceptable only when the return value is truly irrelevant:

```go
// Acceptable — gin's c.Error() returns *gin.Error for chaining, not a failure signal
_ = c.Error(err)

// Never do this — real errors are lost
result, _ := s.repo.Create(ctx, params)
```

---

## Interfaces

### Depend on interfaces, not concrete types

Every constructor in goclarc takes an interface and returns an interface:

```go
// Handler depends on Service interface — not *service
func NewHandler(service Service) *Handler {
    return &Handler{service: service}
}

// Service constructor returns the Service interface — not *service
func NewService(repo Repository) Service {
    return &service{repo: repo}
}
```

This enables swapping implementations in tests without any framework:

```go
type mockRepo struct{}
func (m *mockRepo) Create(ctx context.Context, p CreateParams) (*Entity, error) { ... }

svc := NewService(&mockRepo{})
```

### Keep interfaces small

Prefer interfaces with one or two methods. The generator produces five-method `Repository` and `Service` interfaces because they model a domain's full CRUD contract. For your own abstractions, start smaller:

```go
// Good — one job
type EmailSender interface {
    Send(ctx context.Context, to, subject, body string) error
}

// Reconsider — too broad, hard to mock, hard to substitute
type ExternalService interface {
    Send(...) error
    Receive(...) error
    Authenticate(...) error
    Subscribe(...) error
}
```

### Compile-time interface checks

Add a blank assignment where it matters to catch drift between interface and implementation at compile time, not at runtime:

```go
var _ Repository = (*repository)(nil)
var _ Service    = (*service)(nil)
```

Place these directly below the struct definition in the generated file.

---

## Functions and Methods

### Multiple return values

Go functions return multiple values. The idiomatic form is `(value, error)`:

```go
func (r *repository) GetByID(ctx context.Context, id string) (*Entity, error) { ... }
```

For functions that return only a boolean indicator alongside the value, use the "comma ok" idiom:

```go
value, ok := myMap[key]
if !ok {
    // key not present
}
```

### Named return values — use sparingly

Named returns are useful when the function is short and the names serve as documentation:

```go
func split(sum int) (x, y int) {
    x = sum * 4 / 9
    y = sum - x
    return
}
```

In the goclarc generated code, named returns are not used — they add complexity without benefit for CRUD operations.

### `defer` for guaranteed cleanup

`defer` runs immediately before the enclosing function returns, regardless of path. Use it for cleanup whenever you acquire a resource:

```go
func (r *repository) List(ctx context.Context) ([]*Entity, error) {
    rows, err := r.pool.Query(ctx, `SELECT ...`)
    if err != nil {
        return nil, fmt.Errorf("user.repository.List: %w", err)
    }
    defer rows.Close()   // guaranteed to run even if scan fails

    var entities []*Entity
    for rows.Next() {
        entity, err := scanEntity(rows)
        if err != nil {
            return nil, fmt.Errorf("user.repository.List scan: %w", err)
        }
        entities = append(entities, entity)
    }
    return entities, nil
}
```

Also check `rows.Err()` after the loop — driver errors during iteration surface there:

```go
if err := rows.Err(); err != nil {
    return nil, fmt.Errorf("user.repository.List rows: %w", err)
}
```

### Pointer vs value receivers — pick one and stay consistent

- **Pointer receiver** (`*T`): use when the method modifies the receiver, the struct is large, or you need consistency with other methods on the type.
- **Value receiver** (`T`): use when the method reads only and the type is small.

All generated entity/repository/service methods use pointer receivers for consistency. When you add methods, keep the same receiver type throughout:

```go
// All on *Entity — consistent pointer receivers
func (e *Entity) ToView() View { ... }
func (e *Entity) IsActive() bool { ... }   // your addition — keep *Entity
```

---

## Data

### Zero values — design types to be useful at zero

Go initialises all variables to their zero value. Good type design makes the zero value meaningful:

```go
// Zero value of *service is nil — not useful, so we always construct with NewService
// Zero value of []Entity is nil — useful as "empty list"
var entities []Entity   // nil, but append() works on nil slices

// Zero value of sync.Mutex is unlocked — immediately usable
var mu sync.Mutex
mu.Lock()
```

### `make` vs `new`

- **`make`** creates slices, maps, and channels with their internal data structures initialised.
- **`new`** allocates a zeroed value and returns a pointer to it.

```go
// Slices — use make when you know the length upfront
views := make([]View, len(entities))   // avoids repeated reallocation

// Maps — always use make before writing
updates := make(map[string]any)

// The generated List handler uses make with a known capacity:
views := make([]View, len(entities))
for i, e := range entities {
    views[i] = e.ToView()
}
```

### `append` and growing slices

When length is unknown upfront, start with `nil` and let `append` handle growth:

```go
var entities []*Entity
for rows.Next() {
    entity, err := scanEntity(rows)
    if err != nil {
        return nil, err
    }
    entities = append(entities, entity)
}
```

### `any` instead of `interface{}`

Since Go 1.18, `any` is the preferred alias for `interface{}`. All generated code uses `any`:

```go
// Good
func Scan(dest ...any) error { ... }
updates := map[string]any{}

// Old style — avoid in new code
func Scan(dest ...interface{}) error { ... }
updates := map[string]interface{}{}
```

### Maps — the "comma ok" idiom

Always use the two-value form when a key might be absent:

```go
value, ok := m[key]
if !ok {
    // key is not in the map
}

// Delete is safe even when key is absent
delete(m, key)
```

---

## Context

### `context.Context` is always first

Every function that does I/O or calls another service takes `context.Context` as its first argument:

```go
func (s *service) Create(ctx context.Context, req CreateRequest) (*Entity, error)
func (r *repository) Create(ctx context.Context, p CreateParams) (*Entity, error)
```

Pass it through — never store it in a struct, never create a new background context mid-request:

```go
// Good — context flows from HTTP request through service to repository
entity, err := h.service.Create(c.Request.Context(), req)

// Bad — breaks cancellation and deadline propagation
entity, err := h.service.Create(context.Background(), req)
```

Context is how timeouts, cancellation, and request-scoped values propagate. If a database call is slow and the client disconnects, `pgx` will cancel the query because the request context is cancelled.

---

## Concurrency

### Goroutines are cheap — but always know when they end

The generated `main.go` launches exactly one goroutine: the HTTP server. Everything else is synchronous.

```go
go func() {
    logger.Info("server started", zap.Int("port", cfg.Port))
    if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
        logger.Fatal("listen", zap.Error(err))
    }
}()
```

When you add goroutines to your service layer, use a `WaitGroup` or channel to ensure the goroutine has finished before the function returns.

### Graceful shutdown pattern

The generated `main.go` uses signal-based graceful shutdown:

```go
ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
defer stop()

// ... start server ...

<-ctx.Done()                                          // block until signal
shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
if err := srv.Shutdown(shutCtx); err != nil {
    logger.Error("shutdown", zap.Error(err))
}
```

This gives in-flight requests up to 10 seconds to complete. When you add background workers, hook them into the same context so they also shut down cleanly.

### Channels for signalling

Prefer channels over shared state for coordination between goroutines. Use `struct{}` channels for pure signals:

```go
done := make(chan struct{})
go func() {
    doWork()
    close(done)
}()
<-done
```

Use buffered channels as semaphores to limit concurrency:

```go
const maxConcurrent = 10
sem := make(chan struct{}, maxConcurrent)

for _, job := range jobs {
    sem <- struct{}{}   // acquire
    go func(j Job) {
        defer func() { <-sem }()   // release
        process(j)
    }(job)
}
```

### "Share memory by communicating"

The Go mantra: instead of sharing a variable protected by a mutex, pass ownership through a channel. When you do need a mutex (e.g., a shared cache), keep it as close to the data as possible and document the lock order.

---

## Embedding

Embedding promotes methods from one type into another without inheritance. Use it to compose behaviour rather than copy-paste:

```go
// Embed zap.Logger in a domain logger to extend it
type DomainLogger struct {
    *zap.Logger
    module string
}

func (l *DomainLogger) Info(msg string, fields ...zap.Field) {
    l.Logger.Info(msg, append(fields, zap.String("module", l.module))...)
}
```

Do not use embedding to work around missing interface methods — that is the sign of an interface that is too wide.

---

## Panic and Recover

### Never use `panic` for expected errors

`panic` is for truly unrecoverable situations — invariants that should never be violated. Library and application code should return `error` instead.

```go
// Bad — callers cannot handle this
func MustParseUUID(s string) uuid.UUID {
    id, err := uuid.Parse(s)
    if err != nil {
        panic(err)  // one bad input kills the whole server
    }
    return id
}

// Good — callers decide how to handle failure
func ParseUUID(s string) (uuid.UUID, error) {
    return uuid.Parse(s)
}
```

### Recover only at boundaries

If you launch goroutines that might panic (e.g., third-party callbacks), recover at the goroutine boundary and convert the panic to a logged error:

```go
func safelyRun(work func()) {
    defer func() {
        if r := recover(); r != nil {
            log.Error("recovered panic", zap.Any("panic", r))
        }
    }()
    work()
}
```

The Gin framework's `Recovery()` middleware does this for HTTP handlers. The generated project does not include it by default — add `r.Use(gin.Recovery())` in `main.go` if you want automatic panic recovery on handlers.

---

## Blank Identifier

### Signalling intentional discard

Use `_` to make it explicit that a return value is intentionally unused. This communicates intent and prevents future readers from wondering if it was an oversight.

```go
// Gin's c.Error() returns *gin.Error for chaining — the return is not meaningful here
_ = c.Error(err)

// Import for side effects only (registers drivers, init() hooks)
import _ "net/http/pprof"
```

### Compile-time interface satisfaction

The `_` pattern checks that a type satisfies an interface without allocating:

```go
// Fails at compile time if *repository no longer satisfies Repository
var _ Repository = (*repository)(nil)
```

---

## Go 1.26 Patterns

goclarc targets Go 1.26. These are the relevant additions to be aware of when extending generated code:

**`any` is now universal** — Go 1.18 introduced `any` as an alias for `interface{}`, and 1.26's `go fix` modernisers will automatically rewrite `interface{}` to `any` in your code. goclarc generates `any` everywhere.

**`io.ReadAll` is 2× faster** — if your service layer reads request payloads or external responses manually, prefer `io.ReadAll` over manual buffer management.

**Green Tea GC (default)** — the new garbage collector reduces GC overhead by 10–40% with no code changes. High-throughput handlers benefit automatically. Don't fight the GC with `sync.Pool` until you have profiler evidence.

**`reflect` iterators** — `Type.Fields()` and `Type.Methods()` now return iterators. If you build generic validation or serialisation utilities on top of generated entities, prefer the iterator form.

**`new(T)` accepts expressions** — Go 1.26 allows `new(T{field: value})` for immediate initialisation. Useful when you construct pointer-to-struct values in one line.

**Run `go fix ./...`** after upgrading to apply all official modernisers to your codebase automatically.

---

## Quick Reference

| Pattern | Rule |
|---|---|
| Package names | Lowercase, single-word, no underscores |
| Multi-word identifiers | `MixedCaps` / `mixedCaps` — never underscore |
| Acronyms | All-caps: `UserID`, `HTTPHandler`, `APIKey` |
| Getters | `Owner()` not `GetOwner()` |
| Single-method interfaces | Method + "-er": `Reader`, `Writer`, `Scanner` |
| Error return position | Always last: `(T, error)` |
| Error wrapping | `fmt.Errorf("layer.Method: %w", err)` |
| Not-found detection | `errors.Is(err, ErrNotFound)` — never string match |
| Context | First argument, always propagated |
| `interface{}` | Use `any` instead |
| Cleanup | `defer resource.Close()` immediately after acquiring |
| Goroutines | Know when they end; use context for cancellation |
| Panic | Only for unrecoverable invariants; never in library code |
