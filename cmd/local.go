package cmd

import (
	"fmt"
	"kver/internal/plugin"
	"os"

	"github.com/spf13/cobra"
)

var localCmd = &cobra.Command{
	Use:   "local <lang> <version>",
	Short: "Set project local version for a language (.kver file)",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		lang := args[0]
		version := args[1]
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Printf("[kver] Failed to get current directory: %v\n", err)
			os.Exit(1)
		}
		p, ok := plugin.Get(lang)
		if !ok {
			fmt.Printf("[kver] Language not supported: %s\n", lang)
			os.Exit(1)
		}
		if err := p.Local(version, cwd); err != nil {
			fmt.Printf("[kver] Local failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("[kver] Local %s version set to %s in %s.\n", lang, version, cwd)
	},
}

func init() {
	rootCmd.AddCommand(localCmd)
}
