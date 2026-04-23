// Package cmd wires the goclarc CLI commands.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "goclarc",
	Short: "Go Clean Architecture scaffolding CLI",
	Long: `goclarc — scaffold production-ready Go APIs following Clean Architecture.

  goclarc new my-api                                         initialise a project
  goclarc module user --db postgres --schema schemas/user.yaml   generate a module`,
}

// Execute is the entry point called by main.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(newCmd)
	rootCmd.AddCommand(moduleCmd)
}
