package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/deBilla/goclarc/internal/assets"
	"github.com/deBilla/goclarc/internal/gen"
	"github.com/deBilla/goclarc/internal/schema"
)

var moduleCmd = &cobra.Command{
	Use:   "module [name]",
	Short: "Generate a typed Clean Architecture module",
	Args:  cobra.ExactArgs(1),
	RunE:  runModule,
}

var (
	moduleSchema   string
	moduleDB       string
	moduleOutDir   string
	moduleQueryDir string
	moduleForce    bool
	moduleDryRun   bool
	moduleModPath  string
)

func init() {
	moduleCmd.Flags().StringVarP(&moduleSchema, "schema", "s", "", "path to schema YAML file (required)")
	moduleCmd.Flags().StringVar(&moduleDB, "db", "postgres", "database adapter: postgres | mongo | rtdb")
	moduleCmd.Flags().StringVarP(&moduleOutDir, "out-dir", "o", "", "output directory (default: internal/modules/<name>)")
	moduleCmd.Flags().StringVar(&moduleQueryDir, "query-dir", "schemas/queries", "directory for generated .sql file (postgres only)")
	moduleCmd.Flags().BoolVarP(&moduleForce, "force", "f", false, "overwrite existing files")
	moduleCmd.Flags().BoolVar(&moduleDryRun, "dry-run", false, "print generated files to stdout without writing")
	moduleCmd.Flags().StringVar(&moduleModPath, "module-path", "", "Go module path (detected from go.mod if omitted)")
	_ = moduleCmd.MarkFlagRequired("schema")
}

func runModule(cmd *cobra.Command, args []string) error {
	modName := args[0]

	s, err := schema.Parse(moduleSchema)
	if err != nil {
		return err
	}

	adapter := schema.AdapterFor(moduleDB)

	modPath := moduleModPath
	if modPath == "" {
		modPath, err = detectModulePath()
		if err != nil {
			return fmt.Errorf("could not detect Go module path: %w (set --module-path explicitly)", err)
		}
	}

	ctx := gen.Build(s, adapter, modPath, modName)

	outDir := moduleOutDir
	if outDir == "" {
		outDir = filepath.Join("internal", "modules", modName)
	}

	tmpl, err := parseModuleTemplates()
	if err != nil {
		return err
	}

	repoTmpl := "repository." + moduleDB + ".go.tmpl"

	files := []gen.File{
		{TemplateName: "entity.go.tmpl", OutputPath: filepath.Join(outDir, "entity.go")},
		{TemplateName: "dto.go.tmpl", OutputPath: filepath.Join(outDir, "dto.go")},
		{TemplateName: repoTmpl, OutputPath: filepath.Join(outDir, "repository.go")},
		{TemplateName: "service.go.tmpl", OutputPath: filepath.Join(outDir, "service.go")},
		{TemplateName: "handler.go.tmpl", OutputPath: filepath.Join(outDir, "handler.go")},
		{TemplateName: "routes.go.tmpl", OutputPath: filepath.Join(outDir, "routes.go")},
	}

	if moduleDB == "postgres" {
		files = append(files, gen.File{
			TemplateName: "query.sql.tmpl",
			OutputPath:   filepath.Join(moduleQueryDir, modName+"s.sql"),
		})
	}

	opts := gen.RenderOptions{DryRun: moduleDryRun, Force: moduleForce}

	if !moduleDryRun {
		fmt.Printf("Generating module %q (db: %s) → %s\n", modName, moduleDB, outDir)
	}

	return gen.RenderFiles(tmpl, ctx, files, opts)
}

func parseModuleTemplates() (*template.Template, error) {
	tmpl := template.New("")
	sub, err := fs.Sub(assets.ModuleFS, "templates/module")
	if err != nil {
		return nil, fmt.Errorf("sub module templates: %w", err)
	}
	entries, err := fs.ReadDir(sub, ".")
	if err != nil {
		return nil, fmt.Errorf("read embedded module templates: %w", err)
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

func detectModulePath() (string, error) {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		return "", err
	}
	for _, line := range splitLines(data) {
		if len(line) > 7 && line[:7] == "module " {
			return line[7:], nil
		}
	}
	return "", fmt.Errorf("module directive not found in go.mod")
}

func splitLines(data []byte) []string {
	var lines []string
	start := 0
	for i, b := range data {
		if b == '\n' {
			lines = append(lines, string(data[start:i]))
			start = i + 1
		}
	}
	if start < len(data) {
		lines = append(lines, string(data[start:]))
	}
	return lines
}
