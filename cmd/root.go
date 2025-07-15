package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	// 注册插件
	_ "kver/plugins/go"
	_ "kver/plugins/python"
	_ "kver/plugins/ruby"
	_ "kver/plugins/nodejs"
)

// KverVersion 由构建时 -ldflags 注入，默认 unknown
var KverVersion = "unknown"

var rootCmd = &cobra.Command{
	Use:   "kver",
	Short: "kver is a cross-language version manager",
	Long:  "kver manages multiple versions of programming languages (Go, Python, Node.js, Ruby, etc.)",
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show kver version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("kver version", KverVersion)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(versionCmd)
	// 这里可以添加 install/uninstall/list 等子命令
}
