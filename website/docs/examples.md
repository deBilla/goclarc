---
sidebar_position: 9
---

# Examples

## Example 1: Blog API (Postgres + MongoDB)

A blog API with two modules — `post` on PostgreSQL for structured queries and `comment` on MongoDB for flexible document storage.

### Project Setup

```bash
goclarc new blog-api --module-path github.com/you/blog-api
cd blog-api
mkdir schemas
```

### Post Schema (Postgres)

```yaml title="schemas/post.yaml"
module: post
table: posts

fields:
  - name: id
    type: uuid
    primary: true
    auto: true
  - name: title
    type: string
    required: true
  - name: slug
    type: string
    required: true
  - name: body
    type: string
    required: true
  - name: author_id
    type: uuid
    required: true
  - name: published
    type: bool
  - name: tags
    type: string[]
  - name: created_at
    type: timestamp
    auto: true
  - name: updated_at
    type: timestamp
    nullable: true
    auto: true
```

```bash
goclarc module post --db postgres --schema schemas/post.yaml
```

### Comment Schema (MongoDB)

```yaml title="schemas/comment.yaml"
module: comment
table: comments

fields:
  - name: id
    type: uuid
    primary: true
    auto: true
  - name: post_id
    type: string
    required: true
  - name: author_id
    type: string
    required: true
  - name: body
    type: string
    required: true
  - name: metadata
    type: json
    nullable: true
  - name: created_at
    type: timestamp
    auto: true
```

```bash
goclarc module comment --db mongo --schema schemas/comment.yaml
```

### Wire in main.go

```go
// Postgres — posts
pool, _ := pgxpool.New(ctx, cfg.DatabaseURL)
postRepo    := post.NewRepository(pool)
postService := post.NewService(postRepo)
postHandler := post.NewHandler(postService)
post.RegisterRoutes(v1, postHandler, middleware.Auth())

// MongoDB — comments
mongoClient, _ := mongo.Connect(options.Client().ApplyURI(cfg.MongoURI))
db := mongoClient.Database("blog-api")
commentRepo    := comment.NewRepository(db)
commentService := comment.NewService(commentRepo)
commentHandler := comment.NewHandler(commentService)
comment.RegisterRoutes(v1, commentHandler, middleware.Auth())
```

---

## Example 2: Adding Business Logic to a Generated Service

The generated `service.go` stubs just delegate to the repository. Here's how to extend `Create` with real logic:

```go title="internal/modules/post/service.go (after generation)"
type service struct {
    repo   Repository
    mailer Mailer   // add your own dependencies
}

func NewService(repo Repository, mailer Mailer) Service {
    return &service{repo: repo, mailer: mailer}
}

func (s *service) Create(ctx context.Context, req CreateRequest) (*Entity, error) {
    // Validate uniqueness
    existing, _ := s.repo.GetBySlug(ctx, req.Slug)
    if existing != nil {
        return nil, fmt.Errorf("post.service.Create: slug already exists")
    }

    entity, err := s.repo.Create(ctx, req.ToCreateParams())
    if err != nil {
        return nil, fmt.Errorf("post.service.Create: %w", err)
    }

    // Send notification
    _ = s.mailer.Send(ctx, "New post published: "+entity.Title)

    return entity, nil
}
```

---

## Example 3: User Preferences (Firebase RTDB)

```yaml title="schemas/preference.yaml"
module: preference
table: preferences

fields:
  - name: id
    type: uuid
    primary: true
    auto: true
  - name: user_id
    type: string
    required: true
  - name: theme
    type: string
  - name: notifications_enabled
    type: bool
  - name: settings
    type: json
    nullable: true
  - name: updated_at
    type: timestamp
    auto: true
```

```bash
goclarc module preference --db rtdb --schema schemas/preference.yaml
```

```go
// Wire in main.go
app, _ := firebase.NewApp(ctx, nil)
dbClient, _ := app.DatabaseWithURL(ctx, cfg.FirebaseDBURL)
preferenceRepo    := preference.NewRepository(dbClient)
preferenceService := preference.NewService(preferenceRepo)
preferenceHandler := preference.NewHandler(preferenceService)
preference.RegisterRoutes(v1, preferenceHandler, middleware.Auth())
```

---

## Example 4: Dry-Run Before Committing

Always preview your module before writing to disk, especially on existing projects:

```bash
goclarc module user \
  --db postgres \
  --schema schemas/user.yaml \
  --dry-run 2>&1 | less
```

Review the output, then run without `--dry-run` to write the files:

```bash
goclarc module user --db postgres --schema schemas/user.yaml
```
