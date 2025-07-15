package cmd

import (
	"fmt"
	"kver/internal/plugin"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var activateCmd = &cobra.Command{
	Use:   "activate [<lang>]",
	Short: "Output shell code to activate current language version(s)",
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		langs := []string{}
		if len(args) == 1 {
			langs = []string{args[0]}
		} else {
			for lang := range plugin.All() {
				langs = append(langs, lang)
			}
		}
		cwd, _ := os.Getwd()
		localKver := filepath.Join(cwd, ".kver")
		localMap := map[string]string{}
		if data, err := os.ReadFile(localKver); err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				if strings.Contains(line, "=") {
					parts := strings.SplitN(line, "=", 2)
					lang := strings.TrimSpace(parts[0])
					ver := strings.TrimSpace(parts[1])
					localMap[lang] = ver
				}
			}
		}
		home, _ := os.UserHomeDir()
		globalDir := filepath.Join(home, ".kver", "versions")
		for _, lang := range langs {
			ver := ""
			if v, ok := localMap[lang]; ok {
				ver = v
			} else {
				link := filepath.Join(globalDir, lang)
				if dest, err := os.Readlink(link); err == nil {
					ver = filepath.Base(dest)
				}
			}
			if ver != "" {
				p, ok := plugin.Get(lang)
				if ok {
					if shell, ok := p.(interface{ ActivateShell(version string) string }); ok {
						fmt.Print(shell.ActivateShell(ver))
					} else {
						// 默认输出 PATH 方案
						home, _ := os.UserHomeDir()
						bin := filepath.Join(home, ".kver", "languages", lang, ver, "bin")
						fmt.Printf("export PATH=\"%s:$PATH\"\n", bin)
					}
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(activateCmd)
}
