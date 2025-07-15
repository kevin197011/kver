package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var currentCmd = &cobra.Command{
	Use:   "current [<lang>]",
	Short: "Show current active version(s)",
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		langs := []string{}
		if len(args) == 1 {
			langs = []string{args[0]}
		} else {
			// 直接从 env.d/*.sh 文件名获取所有已激活语言
			home, _ := os.UserHomeDir()
			envDir := filepath.Join(home, ".kver", "env.d")
			entries, err := os.ReadDir(envDir)
			if err == nil {
				for _, e := range entries {
					if strings.HasSuffix(e.Name(), ".sh") {
						lang := strings.TrimSuffix(e.Name(), ".sh")
						langs = append(langs, lang)
					}
				}
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
		envDir := filepath.Join(home, ".kver", "env.d")
		for _, lang := range langs {
			if v, ok := localMap[lang]; ok {
				fmt.Printf("%s: %s (local)\n", lang, v)
				continue
			}
			envFile := filepath.Join(envDir, lang+".sh")
			data, err := os.ReadFile(envFile)
			if err != nil {
				fmt.Printf("%s: (not set)\n", lang)
				continue
			}
			lines := strings.Split(string(data), "\n")
			ver := ""
			for _, line := range lines {
				if strings.Contains(line, "/languages/") {
					// 例: export GOROOT="$HOME/.kver/languages/go/1.21.0"
					parts := strings.Split(line, "/languages/")
					if len(parts) > 1 {
						rest := parts[1]
						restParts := strings.Split(rest, "/")
						if len(restParts) >= 2 {
							ver = restParts[1]
							break
						}
					}
				}
			}
			if ver != "" {
				fmt.Printf("%s: %s (global)\n", lang, ver)
			} else {
				fmt.Printf("%s: (not set)\n", lang)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(currentCmd)
}
