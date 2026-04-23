---
sidebar_position: 2
---

# MongoDB Adapter

```bash
goclarc module post --db mongo --schema schemas/post.yaml
```

## What Gets Generated

```
internal/modules/post/repository.go   ← mongo-driver/v2 implementation
```

No SQL file is generated for MongoDB.

## Repository Pattern

The generated repository uses [mongo-driver/v2](https://pkg.go.dev/go.mongodb.org/mongo-driver/v2):

```go
type repository struct {
    coll *mongo.Collection
}

func NewRepository(db *mongo.Database) Repository {
    return &repository{coll: db.Collection("posts")}
}

func (r *repository) GetByID(ctx context.Context, id string) (*Entity, error) {
    oid, err := bson.ObjectIDFromHex(id)
    if err != nil {
        return nil, fmt.Errorf("post not found")
    }
    var doc bsonPost
    if err := r.coll.FindOne(ctx, bson.M{"_id": oid}).Decode(&doc); err != nil {
        if errors.Is(err, mongo.ErrNoDocuments) {
            return nil, fmt.Errorf("post not found")
        }
        return nil, fmt.Errorf("post.repository.GetByID: %w", err)
    }
    return bsonToEntity(&doc), nil
}
```

## Internal Document Type

The repository uses an internal `bsonPost` struct that maps the primary key to `bson.ObjectID`:

```go
type bsonPost struct {
    ID      bson.ObjectID `bson:"_id,omitempty"`
    Title   string        `bson:"title"`
    Content string        `bson:"content"`
}
```

The `bsonToEntity()` function converts `ObjectID.Hex()` back to the `string` ID in your Entity.

## JSON Fields

For `type: json` fields, the MongoDB adapter uses `bson.M`:

```go
// Entity
Metadata bson.M

// View
Metadata bson.M
```

## Partial Updates

The `Update` method builds a `$set` document from non-nil fields only:

```go
set := bson.M{}
if p.Title != nil { set["title"] = p.Title }
if p.Content != nil { set["content"] = p.Content }
r.coll.FindOneAndUpdate(ctx, bson.M{"_id": oid}, bson.M{"$set": set}, opts)
```

## Required Imports

```go
import (
    "go.mongodb.org/mongo-driver/v2/bson"
    "go.mongodb.org/mongo-driver/v2/mongo"
    "go.mongodb.org/mongo-driver/v2/mongo/options"
)
```

```bash
go get go.mongodb.org/mongo-driver/v2
```

## Connection Setup

```go
client, err := mongo.Connect(options.Client().ApplyURI(cfg.MongoURI))
if err != nil {
    logger.Fatal("mongo", zap.Error(err))
}
defer client.Disconnect(ctx)

db := client.Database("my-api")

postRepo    := post.NewRepository(db)
postService := post.NewService(postRepo)
postHandler := post.NewHandler(postService)
post.RegisterRoutes(v1, postHandler, middleware.Auth())
```
