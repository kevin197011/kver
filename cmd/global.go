package cmd

import (
	"fmt"
	"kver/internal/plugin"
	"os"

	"github.com/spf13/cobra"
)

var globalCmd = &cobra.Command{
	Use:   "global <lang> <version>",
	Short: "Set global default version for a language",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		lang := args[0]
		version := args[1]
		p, ok := plugin.Get(lang)
		if !ok {
			fmt.Printf("[kver] Language not supported: %s\n", lang)
			os.Exit(1)
		}
		if err := p.Global(version); err != nil {
			fmt.Printf("[kver] Global failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("[kver] Global %s version set to %s.\n", lang, version)
	},
}

func init() {
	rootCmd.AddCommand(globalCmd)
}
