package cmd

import (
	"fmt"
	"kver/internal/plugin"
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
			if v, ok := localMap[lang]; ok {
				fmt.Printf("%s: %s (local)\n", lang, v)
				continue
			}
			link := filepath.Join(globalDir, lang)
			if dest, err := os.Readlink(link); err == nil {
				ver := filepath.Base(dest)
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
