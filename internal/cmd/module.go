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
	moduleSchema       string
	moduleDB           string
	moduleOutDir       string
	moduleQueryDir     string
	moduleMigrationDir string
	moduleForce        bool
	moduleDryRun       bool
	moduleReset        bool
	moduleModPath      string
	moduleSwagger      bool
	moduleMode         string
	moduleParent       string
)

func init() {
	moduleCmd.Flags().StringVarP(&moduleSchema, "schema", "s", "", "path to schema YAML file (required)")
	moduleCmd.Flags().StringVar(&moduleDB, "db", "postgres", "database adapter: postgres | mongo | rtdb")
	moduleCmd.Flags().StringVarP(&moduleOutDir, "out-dir", "o", "", "output directory (default: internal/modules/<name>)")
	moduleCmd.Flags().StringVar(&moduleQueryDir, "query-dir", "schemas/queries", "directory for generated .sql file (postgres only)")
	moduleCmd.Flags().StringVar(&moduleMigrationDir, "migration-dir", "db/migrations", "directory for generated CREATE TABLE migration (postgres only)")
	moduleCmd.Flags().BoolVarP(&moduleForce, "force", "f", false, "overwrite existing files")
	moduleCmd.Flags().BoolVar(&moduleDryRun, "dry-run", false, "print generated files to stdout without writing")
	moduleCmd.Flags().BoolVar(&moduleReset, "reset", false, "delete all files previously generated for this module")
	moduleCmd.Flags().StringVar(&moduleModPath, "module-path", "", "Go module path (detected from go.mod if omitted)")
	moduleCmd.Flags().BoolVar(&moduleSwagger, "swagger", false, "generate docs/<name>.openapi.yaml spec for this module")
	moduleCmd.Flags().StringVar(&moduleMode, "mode", "", `generation mode: "" full stack (default) | "repo" entity+dto+repo+migration only`)
	moduleCmd.Flags().StringVar(&moduleParent, "parent", "", `parent route prefix for nested resources (e.g. "/workspaces/:workspaceId")`)
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

	if moduleParent != "" {
		if err := validateParentPath(moduleParent); err != nil {
			return err
		}
		ctx.ParentPath = moduleParent
	}

	outDir := moduleOutDir
	if outDir == "" {
		outDir = filepath.Join("internal", "modules", modName)
	}

	files := moduleFiles(ctx, modName, outDir)

	if moduleReset {
		return resetModule(files, outDir)
	}

	tmpl, err := parseModuleTemplates()
	if err != nil {
		return err
	}

	opts := gen.RenderOptions{DryRun: moduleDryRun, Force: moduleForce}

	if !moduleDryRun {
		fmt.Printf("Generating module %q (db: %s) → %s\n", modName, moduleDB, outDir)
	}

	return gen.RenderFiles(tmpl, ctx, files, opts)
}

// moduleFiles returns the list of files to generate for a module.
// When --mode repo is set, only entity.go, dto.go, repository.go, and
// migration.sql (postgres only) are included; service, handler, and routes
// are skipped.
func moduleFiles(ctx gen.TemplateContext, modName, outDir string) []gen.File {
	repoTmpl := "repository." + moduleDB + ".go.tmpl"

	files := []gen.File{
		{TemplateName: "entity.go.tmpl", OutputPath: filepath.Join(outDir, "entity.go")},
		{TemplateName: "dto.go.tmpl", OutputPath: filepath.Join(outDir, "dto.go")},
		{TemplateName: repoTmpl, OutputPath: filepath.Join(outDir, "repository.go")},
	}

	if moduleMode != "repo" {
		files = append(files,
			gen.File{TemplateName: "service.go.tmpl", OutputPath: filepath.Join(outDir, "service.go")},
			gen.File{TemplateName: "handler.go.tmpl", OutputPath: filepath.Join(outDir, "handler.go")},
			gen.File{TemplateName: "routes.go.tmpl", OutputPath: filepath.Join(outDir, "routes.go")},
		)
	}

	if moduleDB == "postgres" {
		if moduleMode != "repo" {
			files = append(files, gen.File{
				TemplateName: "query.sql.tmpl",
				OutputPath:   filepath.Join(moduleQueryDir, modName+"s.sql"),
			})
		}
		files = append(files, gen.File{
			TemplateName: "migration.sql.tmpl",
			OutputPath:   filepath.Join(moduleMigrationDir, "001_create_"+ctx.TableName+".sql"),
		})
	}

	if moduleSwagger {
		files = append(files, gen.File{
			TemplateName: "openapi.yaml.tmpl",
			OutputPath:   filepath.Join("docs", modName+".openapi.yaml"),
		})
	}

	return files
}

// resetModule deletes all generated files for a module and removes the module
// directory if it is empty afterwards.
func resetModule(files []gen.File, outDir string) error {
	fmt.Printf("Resetting module → removing generated files\n")
	for _, f := range files {
		if err := os.Remove(f.OutputPath); err != nil {
			if os.IsNotExist(err) {
				fmt.Printf("  skipped  %s (not found)\n", f.OutputPath)
				continue
			}
			return fmt.Errorf("reset: remove %s: %w", f.OutputPath, err)
		}
		fmt.Printf("  removed  %s\n", f.OutputPath)
	}

	// Remove the module directory if empty.
	entries, err := os.ReadDir(outDir)
	if err == nil && len(entries) == 0 {
		if err := os.Remove(outDir); err == nil {
			fmt.Printf("  removed  %s/\n", outDir)
		}
	}

	return nil
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

// validateParentPath checks that the parent route prefix does not contain a
// ":id" wildcard, which would conflict with the module's own /:id segments and
// cause Gin to panic at startup.
func validateParentPath(parent string) error {
	for _, segment := range strings.Split(parent, "/") {
		if segment == ":id" {
			return fmt.Errorf("--parent %q contains :id which conflicts with the module's own /:id param; use a distinct name (e.g. :workspaceId)", parent)
		}
	}
	return nil
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
