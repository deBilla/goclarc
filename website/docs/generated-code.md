---
sidebar_position: 8
---

# Generated Code Guide

This page explains what each generated file does and what you're expected to modify after generation.

## File Overview

For `goclarc module user --db postgres`, six files are generated:

```
internal/modules/user/
  entity.go
  dto.go
  repository.go
  service.go
  handler.go
  routes.go
```

---

## `entity.go`

Defines the domain model and the public API view.

```go
// Entity is the internal domain representation.
type Entity struct {
    ID        string
    Email     string
    Name      string
    CreatedAt time.Time
}

// View is the JSON API representation.
type View struct {
    ID        string `json:"id"`
    Email     string `json:"email"`
    Name      string `json:"name"`
    CreatedAt string `json:"created_at"`
}

// ToView converts the domain entity to its public API view.
func (e *Entity) ToView() View {
    return View{
        ID:        e.ID,
        Email:     e.Email,
        Name:      e.Name,
        CreatedAt: e.CreatedAt.Format(time.RFC3339),
    }
}
```

**What you might add:** Extra computed fields in `ToView()`, additional view types (e.g., `ToOwnerView()` with extra sensitive fields).

---

## `dto.go`

Defines request structures for Create and Update operations.

```go
// CreateRequest is bound from the POST request body.
type CreateRequest struct {
    Email string `json:"email" binding:"required"`
    Name  string `json:"name"  binding:"required"`
}

// UpdateRequest is bound from the PATCH request body.
// All fields are optional — only non-nil values are applied.
type UpdateRequest struct {
    Email *string `json:"email"`
    Name  *string `json:"name"`
}
```

**What you might add:** Custom validation beyond `binding:"required"`, field transformations before passing to the service.

---

## `repository.go`

Defines the data access contract (interface) and the DB-specific implementation.

```go
// Repository — the interface your service depends on.
type Repository interface {
    Create(ctx context.Context, p CreateParams) (*Entity, error)
    GetByID(ctx context.Context, id string) (*Entity, error)
    List(ctx context.Context) ([]*Entity, error)
    Update(ctx context.Context, p UpdateParams) (*Entity, error)
    Delete(ctx context.Context, id string) error
}

// repository — the concrete implementation (unexported).
type repository struct { q *sqlcdb.Queries }
```

**What you might add:** Additional query methods (e.g., `ListByEmail`, `GetByUserID`), pagination parameters to `List`.

---

## `service.go`

Implements the business logic layer.

```go
// Service — the interface your handler depends on.
type Service interface {
    Create(ctx context.Context, req CreateRequest) (*Entity, error)
    GetByID(ctx context.Context, id string) (*Entity, error)
    List(ctx context.Context) ([]*Entity, error)
    Update(ctx context.Context, id string, req UpdateRequest) (*Entity, error)
    Delete(ctx context.Context, id string) error
}
```

**What you should add:** Real business logic — authorization checks, cascading operations, event emission, email sending, etc. The generated stubs just delegate to the repository.

---

## `handler.go`

Gin HTTP handlers — one function per operation.

```go
func (h *Handler) Create(c *gin.Context) {
    var req CreateRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        _ = c.Error(err)
        return
    }
    entity, err := h.service.Create(c.Request.Context(), req)
    if err != nil {
        _ = c.Error(err)
        return
    }
    c.JSON(http.StatusCreated, gin.H{"success": true, "data": entity.ToView()})
}
```

**Convention:** Handlers never contain business logic. They parse, delegate, and respond. Errors always go to `c.Error(err)` — the global error middleware handles the response.

**What you might add:** Extra header parsing, pagination query params, auth context extraction (e.g., `c.GetString("userID")`).

---

## `routes.go`

Wires all 5 endpoints onto a Gin router group.

```go
func RegisterRoutes(rg *gin.RouterGroup, h *Handler, middlewares ...gin.HandlerFunc) {
    g := rg.Group("/users", middlewares...)
    g.POST("", h.Create)
    g.GET("", h.List)
    g.GET("/:id", h.GetByID)
    g.PATCH("/:id", h.Update)
    g.DELETE("/:id", h.Delete)
}
```

**What you might add:** Additional routes (e.g., bulk operations, nested resources), different middleware per route.

---

## Regenerating

Use `--force` to overwrite existing files after changing a schema:

```bash
goclarc module user --db postgres --schema schemas/user.yaml --force
```

:::caution
`--force` overwrites generated files. Any custom additions (business logic in `service.go`, extra queries in `repository.go`) will be lost. Consider using `--dry-run` first and applying changes manually.
:::
