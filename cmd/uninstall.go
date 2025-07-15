package cmd

import (
	"fmt"
	"kver/internal/plugin"
	"os"

	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall <lang> <version>",
	Short: "Uninstall a language version",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		lang := args[0]
		version := args[1]
		p, ok := plugin.Get(lang)
		if !ok {
			fmt.Printf("[kver] Language not supported: %s\n", lang)
			os.Exit(1)
		}
		if err := p.Uninstall(version); err != nil {
			fmt.Printf("[kver] Uninstall failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("[kver] %s %s uninstalled.\n", lang, version)
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}
