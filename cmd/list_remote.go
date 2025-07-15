package cmd

import (
	"fmt"
	"kver/internal/plugin"
	"os"

	"github.com/spf13/cobra"
)

var listRemoteCmd = &cobra.Command{
	Use:   "list-remote <lang>",
	Short: "List remote available versions of a language",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		lang := args[0]
		p, ok := plugin.Get(lang)
		if !ok {
			fmt.Printf("[kver] Language not supported: %s\n", lang)
			os.Exit(1)
		}
		versions, err := p.ListRemote()
		if err != nil {
			fmt.Printf("[kver] List-remote failed: %v\n", err)
			os.Exit(1)
		}
		for _, v := range versions {
			fmt.Println(v)
		}
	},
}

func init() {
	rootCmd.AddCommand(listRemoteCmd)
}
