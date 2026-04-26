package cmd

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/deBilla/goclarc/internal/assets"
)

var cryptoCmd = &cobra.Command{
	Use:   "crypto",
	Short: "Generate E2EE cryptography boilerplate into an existing project",
	Long: `crypto writes three files into the target directory:

  ecies.go         — X25519 ECDH + HKDF-SHA512 + AES-256-GCM (16-byte nonce)
  jwt.go           — GenerateAccessToken, GenerateRefreshToken, ParseToken (HS256)
  redis_tokens.go  — IssueTokens, RevokeToken, ValidateToken (rt:{userID}:{jti} pattern)`,
	RunE: runCrypto,
}

var (
	cryptoType string
	cryptoOut  string
	cryptoForce bool
)

func init() {
	cryptoCmd.Flags().StringVar(&cryptoType, "type", "ecies", "crypto type: ecies")
	cryptoCmd.Flags().StringVarP(&cryptoOut, "out", "o", "internal/core/crypto", "output directory")
	cryptoCmd.Flags().BoolVarP(&cryptoForce, "force", "f", false, "overwrite existing files")
}

type cryptoContext struct {
	Package string
}

func runCrypto(_ *cobra.Command, _ []string) error {
	outDir := cryptoOut
	ctx := cryptoContext{Package: filepath.Base(outDir)}

	tmpl, err := parseCryptoTemplates()
	if err != nil {
		return err
	}

	outFiles := []struct{ tmpl, out string }{
		{"ecies.go.tmpl", filepath.Join(outDir, "ecies.go")},
		{"jwt.go.tmpl", filepath.Join(outDir, "jwt.go")},
		{"redis_tokens.go.tmpl", filepath.Join(outDir, "redis_tokens.go")},
	}

	fmt.Printf("Generating crypto boilerplate → %s\n", outDir)

	for _, f := range outFiles {
		if err := writeCryptoFile(tmpl, ctx, f.tmpl, f.out); err != nil {
			return err
		}
	}
	return nil
}

func writeCryptoFile(tmpl *template.Template, ctx cryptoContext, tmplName, outPath string) error {
	t := tmpl.Lookup(tmplName)
	if t == nil {
		return fmt.Errorf("crypto: template %q not found", tmplName)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, ctx); err != nil {
		return fmt.Errorf("crypto: execute %q: %w", tmplName, err)
	}

	if !cryptoForce {
		if _, err := os.Stat(outPath); err == nil {
			return fmt.Errorf("crypto: %s already exists (use --force to overwrite)", outPath)
		}
	}

	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return fmt.Errorf("crypto: mkdir: %w", err)
	}

	if err := os.WriteFile(outPath, buf.Bytes(), 0o644); err != nil {
		return fmt.Errorf("crypto: write %s: %w", outPath, err)
	}

	fmt.Printf("  created  %s\n", outPath)
	return nil
}

func parseCryptoTemplates() (*template.Template, error) {
	tmpl := template.New("")
	sub, err := fs.Sub(assets.CryptoFS, "templates/crypto")
	if err != nil {
		return nil, fmt.Errorf("sub crypto templates: %w", err)
	}
	entries, err := fs.ReadDir(sub, ".")
	if err != nil {
		return nil, fmt.Errorf("read embedded crypto templates: %w", err)
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
