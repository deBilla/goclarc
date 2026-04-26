package gen

import (
	"fmt"
	"strings"

	"github.com/deBilla/goclarc/internal/schema"
)

// FieldContext holds all the derived information for one field used in templates.
type FieldContext struct {
	// Raw name from schema (snake_case).
	Name string
	// GoName is the PascalCase struct field name (e.g. "UserID").
	GoName string
	// JSONTag is the snake_case json tag value.
	JSONTag string
	// EntityType is the Go type in the Entity struct.
	EntityType string
	// ViewType is the Go type in the View struct.
	ViewType string
	// CreateType is the Go type in CreateParams.
	CreateType string
	// UpdateType is the Go type in UpdateParams (always pointer).
	UpdateType string
	// ZeroValue is the zero literal for this type.
	ZeroValue string

	IsPrimary         bool
	IsAuto            bool // skipped from Create/Update params
	IsRequired        bool
	IsNullable        bool
	NeedsTimeImport   bool
	NeedsJSONImport   bool
	IsPointerNullable bool
	ViewIsString      bool
	// IsNullableScalar is true when the entity field is a pointer (*T) but the
	// view field is a non-pointer (T), requiring a nil-guarded dereference in ToView.
	IsNullableScalar bool
	// IsLast is true for the final field in the Fields slice — used in templates
	// that need to suppress a trailing comma (e.g. SQL CREATE TABLE).
	IsLast bool
	// BindingTag is the Gin binding tag (e.g. `binding:"required"`).
	BindingTag string
	// OAType is the OpenAPI primitive type (string, integer, number, boolean, object, array).
	OAType string
	// OAFormat is the OpenAPI format qualifier (uuid, int32, int64, float, date-time). Empty when none applies.
	OAFormat string
	// OAItemType is the OpenAPI items type for array fields (e.g. "string" for string[]).
	OAItemType string
}

// TemplateContext is the data passed to every template.
type TemplateContext struct {
	// ModuleName is the lower-case module name (e.g. "user").
	ModuleName string
	// ModuleTitle is the PascalCase module name (e.g. "User").
	ModuleTitle string
	// ModulePlural is the plural kebab-case for routes (e.g. "users").
	ModulePlural string
	// TableName is the DB table/collection/ref name.
	TableName string
	// ModulePath is the Go module path for imports.
	ModulePath string
	// DBAdapter is the database adapter name ("postgres", "mongo", "rtdb").
	DBAdapter string
	// DBImports are the extra imports needed in repository.go.
	DBImports []string

	Fields       []FieldContext
	CreateFields []FieldContext // fields included in CreateParams (non-auto)
	UpdateFields []FieldContext // fields included in UpdateParams (non-auto, non-primary)

	HasTimeImport bool
	HasJSONImport bool
	// HasRequiredCreateFields is true when at least one CreateField has IsRequired=true.
	// Used by openapi.yaml.tmpl to emit the required: block.
	HasRequiredCreateFields bool

	// PostgreSQL raw-query helpers (populated only for the postgres adapter).
	InsertColumns      string // "email, name, age"
	InsertPlaceholders string // "$1, $2, $3"
	SelectColumns      string // "id, email, name, age, created_at"
	UpdateSetClause    string // "email = COALESCE($2, email), ..."
}

// Build assembles a TemplateContext from a parsed schema and a DB adapter.
// moduleName overrides the schema's module field when non-empty.
func Build(s *schema.Schema, adapter schema.Adapter, modulePath, moduleName string) TemplateContext {
	name := s.Module
	if moduleName != "" {
		name = moduleName
	}
	ctx := TemplateContext{
		ModuleName:   name,
		ModuleTitle:  ToPascal(name),
		ModulePlural: ToPlural(name),
		TableName:    s.Table,
		ModulePath:   modulePath,
		DBAdapter:    adapter.Name(),
		DBImports:    adapter.DBImports(modulePath),
	}

	for _, f := range s.Fields {
		info := adapter.Map(f.Type, f.Nullable)

		binding := ""
		if f.Required {
			binding = `binding:"required"`
		}

		fc := FieldContext{
			Name:              f.Name,
			GoName:            ToPascal(f.Name),
			JSONTag:           f.Name,
			EntityType:        info.EntityType,
			ViewType:          info.ViewType,
			CreateType:        info.CreateType,
			UpdateType:        info.UpdateType,
			ZeroValue:         info.ZeroValue,
			IsPrimary:         f.Primary,
			IsAuto:            f.Auto,
			IsRequired:        f.Required,
			IsNullable:        f.Nullable,
			NeedsTimeImport:   info.NeedsTimeImport,
			NeedsJSONImport:   info.NeedsJSONImport,
			IsPointerNullable: info.IsPointerNullable,
			ViewIsString:      info.ViewIsString,
			IsNullableScalar:  f.Nullable && !info.ViewIsString && !f.Primary,
			BindingTag:        binding,
			OAType:            info.OAType,
			OAFormat:          info.OAFormat,
			OAItemType:        info.OAItemType,
		}

		ctx.Fields = append(ctx.Fields, fc)
		// Mark previous field as no longer last.
		if len(ctx.Fields) > 1 {
			ctx.Fields[len(ctx.Fields)-2].IsLast = false
		}
		ctx.Fields[len(ctx.Fields)-1].IsLast = true

		if info.NeedsTimeImport {
			ctx.HasTimeImport = true
		}
		if info.NeedsJSONImport {
			ctx.HasJSONImport = true
		}

		if !f.Auto {
			ctx.CreateFields = append(ctx.CreateFields, fc)
		}
		if !f.Auto && !f.Primary {
			ctx.UpdateFields = append(ctx.UpdateFields, fc)
		}
	}

	for _, fc := range ctx.CreateFields {
		if fc.IsRequired {
			ctx.HasRequiredCreateFields = true
			break
		}
	}

	if adapter.Name() == "postgres" {
		var insertCols, insertPlaceholders, selectCols, updateSet []string
		for i, f := range ctx.CreateFields {
			insertCols = append(insertCols, f.Name)
			insertPlaceholders = append(insertPlaceholders, fmt.Sprintf("$%d", i+1))
		}
		for _, f := range ctx.Fields {
			selectCols = append(selectCols, f.Name)
		}
		for i, f := range ctx.UpdateFields {
			updateSet = append(updateSet, fmt.Sprintf("%s = COALESCE($%d, %s)", f.Name, i+2, f.Name))
		}
		ctx.InsertColumns = strings.Join(insertCols, ", ")
		ctx.InsertPlaceholders = strings.Join(insertPlaceholders, ", ")
		ctx.SelectColumns = strings.Join(selectCols, ", ")
		ctx.UpdateSetClause = strings.Join(updateSet, ",\n\t\t  ")
	}

	return ctx
}
