package cmd

import (
	"fmt"
	"kver/internal/plugin"
	"os"

	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install <lang> <version>",
	Short: "Download and install a language version",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		lang := args[0]
		version := args[1]
		p, ok := plugin.Get(lang)
		if !ok {
			fmt.Printf("[kver] Language not supported: %s\n", lang)
			os.Exit(1)
		}
		if err := p.Install(version); err != nil {
			fmt.Printf("[kver] Install failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("[kver] %s %s installed successfully.\n", lang, version)
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
