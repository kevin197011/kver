package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	// 注册插件
	_ "kver/plugins/ruby"
)

var rootCmd = &cobra.Command{
	Use:   "kver",
	Short: "kver is a cross-language version manager",
	Long:  "kver manages multiple versions of programming languages (Go, Python, Node.js, Ruby, etc.)",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// 这里可以添加 install/uninstall/list 等子命令
}
