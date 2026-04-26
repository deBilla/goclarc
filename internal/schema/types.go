// Package schema handles parsing and type-mapping of goclarc schema files.
package schema

// GoTypeInfo holds all the type information needed to render a single field
// across all generated files for a given database adapter.
type GoTypeInfo struct {
	// EntityType is the Go type used in the Entity struct (e.g. "string", "*time.Time").
	EntityType string
	// ViewType is the Go type used in the View struct (usually same as EntityType or "string" for timestamps).
	ViewType string
	// CreateType is the Go type used in CreateParams (same as EntityType for non-nullable, may differ for nullable).
	CreateType string
	// UpdateType is the Go type used in UpdateParams (always a pointer to allow partial updates).
	UpdateType string
	// ZeroValue is the zero value literal for this type (used in stub implementations).
	ZeroValue string
	// NeedsTimeImport signals that "time" must be imported in entity.go.
	NeedsTimeImport bool
	// NeedsJSONImport signals that "encoding/json" must be imported.
	NeedsJSONImport bool
	// IsPointerNullable means a nullable timestamp uses the "if Valid { t := ...; e.F = &t }" pattern.
	IsPointerNullable bool
	// ViewIsString means the View field should be rendered as a string (e.g. time.RFC3339).
	ViewIsString bool
	// OAType is the OpenAPI primitive type (string, integer, number, boolean, object, array).
	OAType string
	// OAFormat is the OpenAPI format qualifier (uuid, int32, int64, float, date-time). Empty when none applies.
	OAFormat string
	// OAItemType is the OpenAPI items type for array fields (e.g. "string" for string[]).
	OAItemType string
}

// Adapter maps schema field types to database-specific Go type information.
type Adapter interface {
	Map(fieldType string, nullable bool) GoTypeInfo
	// DBImports returns extra imports needed in repository.go for this adapter.
	DBImports(modulePath string) []string
	// Name returns the adapter identifier ("postgres", "mongo", "rtdb").
	Name() string
}

// --- PostgreSQL adapter ---

type postgresAdapter struct{}

// PostgresAdapter returns the PostgreSQL (sqlc + pgx/v5) type adapter.
func PostgresAdapter() Adapter { return &postgresAdapter{} }

func (a *postgresAdapter) Name() string { return "postgres" }

func (a *postgresAdapter) DBImports(modulePath string) []string {
	return []string{
		"errors",
		"fmt",
		"github.com/jackc/pgx/v5",
		"github.com/jackc/pgx/v5/pgxpool",
		modulePath + "/internal/core/pghelper",
		modulePath + "/internal/core/sqlcdb",
	}
}

func (a *postgresAdapter) Map(fieldType string, nullable bool) GoTypeInfo {
	switch fieldType {
	case "uuid":
		return GoTypeInfo{
			EntityType: "string", ViewType: "string",
			CreateType: "string", UpdateType: "*string",
			ZeroValue: `""`,
			OAType: "string", OAFormat: "uuid",
		}
	case "string":
		if nullable {
			return GoTypeInfo{EntityType: "*string", ViewType: "string", CreateType: "*string", UpdateType: "*string", ZeroValue: "nil", OAType: "string"}
		}
		return GoTypeInfo{EntityType: "string", ViewType: "string", CreateType: "string", UpdateType: "*string", ZeroValue: `""`, OAType: "string"}
	case "int":
		if nullable {
			return GoTypeInfo{EntityType: "*int32", ViewType: "int32", CreateType: "*int32", UpdateType: "*int32", ZeroValue: "nil", OAType: "integer", OAFormat: "int32"}
		}
		return GoTypeInfo{EntityType: "int32", ViewType: "int32", CreateType: "int32", UpdateType: "*int32", ZeroValue: "0", OAType: "integer", OAFormat: "int32"}
	case "int64":
		if nullable {
			return GoTypeInfo{EntityType: "*int64", ViewType: "int64", CreateType: "*int64", UpdateType: "*int64", ZeroValue: "nil", OAType: "integer", OAFormat: "int64"}
		}
		return GoTypeInfo{EntityType: "int64", ViewType: "int64", CreateType: "int64", UpdateType: "*int64", ZeroValue: "0", OAType: "integer", OAFormat: "int64"}
	case "float":
		if nullable {
			return GoTypeInfo{EntityType: "*float64", ViewType: "float64", CreateType: "*float64", UpdateType: "*float64", ZeroValue: "nil", OAType: "number", OAFormat: "float"}
		}
		return GoTypeInfo{EntityType: "float64", ViewType: "float64", CreateType: "float64", UpdateType: "*float64", ZeroValue: "0", OAType: "number", OAFormat: "float"}
	case "bool":
		if nullable {
			return GoTypeInfo{EntityType: "*bool", ViewType: "bool", CreateType: "*bool", UpdateType: "*bool", ZeroValue: "nil", OAType: "boolean"}
		}
		return GoTypeInfo{EntityType: "bool", ViewType: "bool", CreateType: "bool", UpdateType: "*bool", ZeroValue: "false", OAType: "boolean"}
	case "timestamp":
		if nullable {
			return GoTypeInfo{
				EntityType: "*time.Time", ViewType: "string",
				CreateType: "*time.Time", UpdateType: "*time.Time",
				ZeroValue: "nil", NeedsTimeImport: true,
				IsPointerNullable: true, ViewIsString: true,
				OAType: "string", OAFormat: "date-time",
			}
		}
		return GoTypeInfo{
			EntityType: "time.Time", ViewType: "string",
			CreateType: "time.Time", UpdateType: "*time.Time",
			ZeroValue: "time.Time{}", NeedsTimeImport: true, ViewIsString: true,
			OAType: "string", OAFormat: "date-time",
		}
	case "json":
		return GoTypeInfo{
			EntityType: "json.RawMessage", ViewType: "json.RawMessage",
			CreateType: "json.RawMessage", UpdateType: "json.RawMessage",
			ZeroValue: "nil", NeedsJSONImport: true,
			OAType: "object",
		}
	case "string[]":
		return GoTypeInfo{EntityType: "[]string", ViewType: "[]string", CreateType: "[]string", UpdateType: "[]string", ZeroValue: "nil", OAType: "array", OAItemType: "string"}
	default:
		return GoTypeInfo{EntityType: "string", ViewType: "string", CreateType: "string", UpdateType: "*string", ZeroValue: `""`, OAType: "string"}
	}
}

// --- MongoDB adapter ---

type mongoAdapter struct{}

// MongoAdapter returns the MongoDB (mongo-driver/v2) type adapter.
func MongoAdapter() Adapter { return &mongoAdapter{} }

func (a *mongoAdapter) Name() string { return "mongo" }

func (a *mongoAdapter) DBImports(modulePath string) []string {
	return []string{
		"fmt",
		"go.mongodb.org/mongo-driver/v2/bson",
		"go.mongodb.org/mongo-driver/v2/mongo",
		modulePath + "/internal/core/errors",
	}
}

func (a *mongoAdapter) Map(fieldType string, nullable bool) GoTypeInfo {
	switch fieldType {
	case "uuid":
		return GoTypeInfo{EntityType: "string", ViewType: "string", CreateType: "string", UpdateType: "*string", ZeroValue: `""`, OAType: "string", OAFormat: "uuid"}
	case "string":
		if nullable {
			return GoTypeInfo{EntityType: "*string", ViewType: "string", CreateType: "*string", UpdateType: "*string", ZeroValue: "nil", OAType: "string"}
		}
		return GoTypeInfo{EntityType: "string", ViewType: "string", CreateType: "string", UpdateType: "*string", ZeroValue: `""`, OAType: "string"}
	case "int":
		if nullable {
			return GoTypeInfo{EntityType: "*int32", ViewType: "int32", CreateType: "*int32", UpdateType: "*int32", ZeroValue: "nil", OAType: "integer", OAFormat: "int32"}
		}
		return GoTypeInfo{EntityType: "int32", ViewType: "int32", CreateType: "int32", UpdateType: "*int32", ZeroValue: "0", OAType: "integer", OAFormat: "int32"}
	case "int64":
		if nullable {
			return GoTypeInfo{EntityType: "*int64", ViewType: "int64", CreateType: "*int64", UpdateType: "*int64", ZeroValue: "nil", OAType: "integer", OAFormat: "int64"}
		}
		return GoTypeInfo{EntityType: "int64", ViewType: "int64", CreateType: "int64", UpdateType: "*int64", ZeroValue: "0", OAType: "integer", OAFormat: "int64"}
	case "float":
		if nullable {
			return GoTypeInfo{EntityType: "*float64", ViewType: "float64", CreateType: "*float64", UpdateType: "*float64", ZeroValue: "nil", OAType: "number", OAFormat: "float"}
		}
		return GoTypeInfo{EntityType: "float64", ViewType: "float64", CreateType: "float64", UpdateType: "*float64", ZeroValue: "0", OAType: "number", OAFormat: "float"}
	case "bool":
		if nullable {
			return GoTypeInfo{EntityType: "*bool", ViewType: "bool", CreateType: "*bool", UpdateType: "*bool", ZeroValue: "nil", OAType: "boolean"}
		}
		return GoTypeInfo{EntityType: "bool", ViewType: "bool", CreateType: "bool", UpdateType: "*bool", ZeroValue: "false", OAType: "boolean"}
	case "timestamp":
		if nullable {
			return GoTypeInfo{
				EntityType: "*time.Time", ViewType: "string",
				CreateType: "*time.Time", UpdateType: "*time.Time",
				ZeroValue: "nil", NeedsTimeImport: true,
				IsPointerNullable: true, ViewIsString: true,
				OAType: "string", OAFormat: "date-time",
			}
		}
		return GoTypeInfo{
			EntityType: "time.Time", ViewType: "string",
			CreateType: "time.Time", UpdateType: "*time.Time",
			ZeroValue: "time.Time{}", NeedsTimeImport: true, ViewIsString: true,
			OAType: "string", OAFormat: "date-time",
		}
	case "json":
		return GoTypeInfo{EntityType: "bson.M", ViewType: "bson.M", CreateType: "bson.M", UpdateType: "bson.M", ZeroValue: "nil", OAType: "object"}
	case "string[]":
		return GoTypeInfo{EntityType: "[]string", ViewType: "[]string", CreateType: "[]string", UpdateType: "[]string", ZeroValue: "nil", OAType: "array", OAItemType: "string"}
	default:
		return GoTypeInfo{EntityType: "string", ViewType: "string", CreateType: "string", UpdateType: "*string", ZeroValue: `""`, OAType: "string"}
	}
}

// --- Firebase RTDB adapter ---

type rtdbAdapter struct{}

// RTDBAdapter returns the Firebase Realtime Database type adapter.
func RTDBAdapter() Adapter { return &rtdbAdapter{} }

func (a *rtdbAdapter) Name() string { return "rtdb" }

func (a *rtdbAdapter) DBImports(modulePath string) []string {
	return []string{
		"fmt",
		"firebase.google.com/go/v4/db",
		modulePath + "/internal/core/errors",
	}
}

func (a *rtdbAdapter) Map(fieldType string, nullable bool) GoTypeInfo {
	switch fieldType {
	case "uuid":
		if nullable {
			return GoTypeInfo{EntityType: "*string", ViewType: "string", CreateType: "*string", UpdateType: "*string", ZeroValue: "nil", OAType: "string", OAFormat: "uuid"}
		}
		return GoTypeInfo{EntityType: "string", ViewType: "string", CreateType: "string", UpdateType: "*string", ZeroValue: `""`, OAType: "string", OAFormat: "uuid"}
	case "string":
		if nullable {
			return GoTypeInfo{EntityType: "*string", ViewType: "string", CreateType: "*string", UpdateType: "*string", ZeroValue: "nil", OAType: "string"}
		}
		return GoTypeInfo{EntityType: "string", ViewType: "string", CreateType: "string", UpdateType: "*string", ZeroValue: `""`, OAType: "string"}
	case "int":
		if nullable {
			return GoTypeInfo{EntityType: "*int", ViewType: "int", CreateType: "*int", UpdateType: "*int", ZeroValue: "nil", OAType: "integer", OAFormat: "int32"}
		}
		return GoTypeInfo{EntityType: "int", ViewType: "int", CreateType: "int", UpdateType: "*int", ZeroValue: "0", OAType: "integer", OAFormat: "int32"}
	case "int64":
		if nullable {
			return GoTypeInfo{EntityType: "*int64", ViewType: "int64", CreateType: "*int64", UpdateType: "*int64", ZeroValue: "nil", OAType: "integer", OAFormat: "int64"}
		}
		return GoTypeInfo{EntityType: "int64", ViewType: "int64", CreateType: "int64", UpdateType: "*int64", ZeroValue: "0", OAType: "integer", OAFormat: "int64"}
	case "float":
		if nullable {
			return GoTypeInfo{EntityType: "*float64", ViewType: "float64", CreateType: "*float64", UpdateType: "*float64", ZeroValue: "nil", OAType: "number", OAFormat: "float"}
		}
		return GoTypeInfo{EntityType: "float64", ViewType: "float64", CreateType: "float64", UpdateType: "*float64", ZeroValue: "0", OAType: "number", OAFormat: "float"}
	case "bool":
		if nullable {
			return GoTypeInfo{EntityType: "*bool", ViewType: "bool", CreateType: "*bool", UpdateType: "*bool", ZeroValue: "nil", OAType: "boolean"}
		}
		return GoTypeInfo{EntityType: "bool", ViewType: "bool", CreateType: "bool", UpdateType: "*bool", ZeroValue: "false", OAType: "boolean"}
	case "timestamp":
		// RTDB stores timestamps as unix milliseconds; exposed as integer in the API.
		return GoTypeInfo{EntityType: "int64", ViewType: "int64", CreateType: "int64", UpdateType: "*int64", ZeroValue: "0", OAType: "integer", OAFormat: "int64"}
	case "json":
		return GoTypeInfo{EntityType: "map[string]any", ViewType: "map[string]any", CreateType: "map[string]any", UpdateType: "map[string]any", ZeroValue: "nil", OAType: "object"}
	case "string[]":
		return GoTypeInfo{EntityType: "[]string", ViewType: "[]string", CreateType: "[]string", UpdateType: "[]string", ZeroValue: "nil", OAType: "array", OAItemType: "string"}
	default:
		return GoTypeInfo{EntityType: "string", ViewType: "string", CreateType: "string", UpdateType: "*string", ZeroValue: `""`, OAType: "string"}
	}
}

// AdapterFor returns the adapter for the given database name.
func AdapterFor(db string) Adapter {
	switch db {
	case "mongo":
		return MongoAdapter()
	case "rtdb":
		return RTDBAdapter()
	default:
		return PostgresAdapter()
	}
}
