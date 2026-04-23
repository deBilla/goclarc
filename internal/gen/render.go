package gen

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

// File describes a single file to be rendered.
type File struct {
	// TemplateName is the key inside the parsed template set (e.g. "entity.go.tmpl").
	TemplateName string
	// OutputPath is the destination path on disk.
	OutputPath string
}

// RenderOptions controls how files are written.
type RenderOptions struct {
	// DryRun prints rendered output to stdout instead of writing files.
	DryRun bool
	// Force allows overwriting existing files.
	Force bool
}

// RenderFiles executes each template against ctx and writes the results.
// tmplFS must be a parsed *template.Template containing all named sub-templates.
func RenderFiles(tmpl *template.Template, ctx TemplateContext, files []File, opts RenderOptions) error {
	for _, f := range files {
		if err := renderOne(tmpl, ctx, f, opts); err != nil {
			return err
		}
	}
	return nil
}

func renderOne(tmpl *template.Template, ctx TemplateContext, f File, opts RenderOptions) error {
	t := tmpl.Lookup(f.TemplateName)
	if t == nil {
		return fmt.Errorf("render: template %q not found", f.TemplateName)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, ctx); err != nil {
		return fmt.Errorf("render: execute %q: %w", f.TemplateName, err)
	}

	if opts.DryRun {
		fmt.Printf("=== %s ===\n%s\n", f.OutputPath, buf.String())
		return nil
	}

	if !opts.Force {
		if _, err := os.Stat(f.OutputPath); err == nil {
			return fmt.Errorf("render: %s already exists (use --force to overwrite)", f.OutputPath)
		}
	}

	if err := os.MkdirAll(filepath.Dir(f.OutputPath), 0o755); err != nil {
		return fmt.Errorf("render: mkdir %s: %w", filepath.Dir(f.OutputPath), err)
	}

	if err := os.WriteFile(f.OutputPath, buf.Bytes(), 0o644); err != nil {
		return fmt.Errorf("render: write %s: %w", f.OutputPath, err)
	}

	fmt.Printf("  created  %s\n", f.OutputPath)
	return nil
}

// ParseGlob wraps template.ParseFS/ParseGlob to load all .tmpl files.
// The returned Template has every file registered by its base name.
func ParseGlob(pattern string) (*template.Template, error) {
	return template.ParseGlob(pattern)
}
