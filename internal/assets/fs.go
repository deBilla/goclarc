// Package assets embeds the goclarc code generation templates.
package assets

import "embed"

// ModuleFS contains all templates under templates/module/.
//
//go:embed templates/module
var ModuleFS embed.FS

// ProjectFS contains all templates under templates/project/ including dotfiles.
//
//go:embed all:templates/project
var ProjectFS embed.FS

// CryptoFS contains all templates under templates/crypto/.
//
//go:embed templates/crypto
var CryptoFS embed.FS
