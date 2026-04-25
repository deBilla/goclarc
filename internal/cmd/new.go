package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/deBilla/goclarc/internal/assets"
)

var newCmd = &cobra.Command{
	Use:   "new [project-name]",
	Short: "Scaffold a new Clean Architecture Go project",
	Args:  cobra.ExactArgs(1),
	RunE:  runNew,
}

var (
	newModulePath  string
	newPort        int
	newDBAdapter   string
)

func init() {
	newCmd.Flags().StringVar(&newModulePath, "module-path", "", "Go module path (default: github.com/<user>/<name>)")
	newCmd.Flags().IntVar(&newPort, "port", 3001, "default HTTP port in generated config")
	newCmd.Flags().StringVar(&newDBAdapter, "db", "postgres", "database adapter: postgres | mongo | rtdb")
}

type projectCtx struct {
	ProjectName string
	ModulePath  string
	Port        int
	DBAdapter   string
}

func runNew(_ *cobra.Command, args []string) error {
	name := args[0]

	switch newDBAdapter {
	case "postgres", "mongo", "rtdb":
	default:
		return fmt.Errorf("unknown database adapter %q (valid: postgres, mongo, rtdb)", newDBAdapter)
	}

	modPath := newModulePath
	if modPath == "" {
		modPath = "github.com/your-user/" + name
	}

	ctx := projectCtx{
		ProjectName: filepath.Base(name),
		ModulePath:  modPath,
		Port:        newPort,
		DBAdapter:   newDBAdapter,
	}

	if _, err := os.Stat(name); err == nil {
		return fmt.Errorf("directory %q already exists", name)
	}

	tmpl, err := parseProjectTemplates()
	if err != nil {
		return err
	}

	fmt.Printf("Creating project %q → ./%s  [db: %s]\n", name, name, newDBAdapter)

	type outFile struct {
		tmplName string
		outPath  string
	}

	projectFiles := []outFile{
		{"go.mod.tmpl", filepath.Join(name, "go.mod")},
		{".gitignore.tmpl", filepath.Join(name, ".gitignore")},
		{".env.example.tmpl", filepath.Join(name, ".env.example")},
		{"main.go.tmpl", filepath.Join(name, "cmd", "api", "main.go")},
		{"config.go.tmpl", filepath.Join(name, "internal", "core", "config", "config.go")},
		{"db." + newDBAdapter + ".go.tmpl", filepath.Join(name, "internal", "core", "db", "db.go")},
		{"errors.go.tmpl", filepath.Join(name, "internal", "core", "errors", "errors.go")},
		{"response.go.tmpl", filepath.Join(name, "internal", "core", "response", "response.go")},
		{"auth.go.tmpl", filepath.Join(name, "internal", "middleware", "auth.go")},
		{"error.go.tmpl", filepath.Join(name, "internal", "middleware", "error.go")},
		{"logger.go.tmpl", filepath.Join(name, "internal", "middleware", "logger.go")},
	}

	for _, f := range projectFiles {
		t := tmpl.Lookup(f.tmplName)
		if t == nil {
			return fmt.Errorf("template %q not found", f.tmplName)
		}
		if err := writeTemplate(t, ctx, f.outPath); err != nil {
			return err
		}
	}

	fmt.Println()
	fmt.Println("Project created. Next steps:")
	fmt.Printf("  cd %s\n", name)
	fmt.Println("  cp .env.example .env   # fill in your secrets")
	fmt.Println("  go mod tidy")
	fmt.Println("  go run ./cmd/api")
	return nil
}

func writeTemplate(t *template.Template, ctx any, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(path), err)
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	defer f.Close()
	if err := t.Execute(f, ctx); err != nil {
		return fmt.Errorf("execute template → %s: %w", path, err)
	}
	fmt.Printf("  created  %s\n", path)
	return nil
}

func parseProjectTemplates() (*template.Template, error) {
	tmpl := template.New("").Funcs(template.FuncMap{
		"toLower": strings.ToLower,
	})
	sub, err := fs.Sub(assets.ProjectFS, "templates/project")
	if err != nil {
		return nil, fmt.Errorf("sub project templates: %w", err)
	}
	entries, err := fs.ReadDir(sub, ".")
	if err != nil {
		return nil, fmt.Errorf("read embedded project templates: %w", err)
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		data, err := fs.ReadFile(sub, e.Name())
		if err != nil {
			return nil, fmt.Errorf("read template %s: %w", e.Name(), err)
		}
		if _, err := tmpl.New(e.Name()).Parse(string(data)); err != nil {
			return nil, fmt.Errorf("parse template %s: %w", e.Name(), err)
		}
	}
	return tmpl, nil
}
