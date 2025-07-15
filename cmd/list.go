package cmd

import (
	"fmt"
	"kver/internal/plugin"
	"os"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list <lang>",
	Short: "List installed versions of a language",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		lang := args[0]
		p, ok := plugin.Get(lang)
		if !ok {
			fmt.Printf("[kver] Language not supported: %s\n", lang)
			os.Exit(1)
		}
		versions, err := p.List()
		if err != nil {
			fmt.Printf("[kver] List failed: %v\n", err)
			os.Exit(1)
		}
		for _, v := range versions {
			fmt.Println(v)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
