// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package nodejs

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"kver/internal/plugin"
)

type NodejsPlugin struct{}

func (n *NodejsPlugin) Name() string { return "nodejs" }

func (n *NodejsPlugin) Install(version string) error {
	home, _ := os.UserHomeDir()
	kverHome := filepath.Join(home, ".kver")
	installDir := filepath.Join(kverHome, "languages", "nodejs", version)

	var installOk bool
	defer func() {
		if !installOk {
			os.RemoveAll(installDir)
		}
	}()

	title := func(s string) {
		fmt.Printf("\n\033[1;36m[kver][nodejs] %s\033[0m\n", s)
	}
	sep := func() {
		fmt.Println("\033[1;34m----------------------------------------\033[0m")
	}

	title("Step 1/3: Download Node.js tarball")
	osStr := runtime.GOOS
	archStr := runtime.GOARCH
	var nodeArch string
	switch archStr {
	case "amd64":
		nodeArch = "x64"
	case "arm64":
		nodeArch = "arm64"
	default:
		return fmt.Errorf("unsupported arch: %s", archStr)
	}
	url := fmt.Sprintf("https://nodejs.org/dist/v%s/node-v%s-%s-%s.tar.gz", version, version, osStr, nodeArch)
	fmt.Printf("[kver][nodejs] Downloading %s\n", url)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("download failed: %s", resp.Status)
	}
	tarball := filepath.Join(os.TempDir(), filepath.Base(url))
	out, err := os.Create(tarball)
	if err != nil {
		return err
	}
	if _, err = io.Copy(out, resp.Body); err != nil {
		out.Close()
		return err
	}
	out.Close()
	sep()

	title("Step 2/3: Extract Node.js tarball to install directory")
	os.RemoveAll(installDir)
	os.MkdirAll(filepath.Dir(installDir), 0755)
	if err := extractTarGz(tarball, filepath.Dir(installDir)); err != nil {
		return err
	}
	// 解压后目录名如 node-vXX，重命名为 installDir
	dirs, _ := os.ReadDir(filepath.Dir(installDir))
	var extractedDir string
	for _, d := range dirs {
		if d.IsDir() && strings.HasPrefix(d.Name(), "node-v") {
			extractedDir = filepath.Join(filepath.Dir(installDir), d.Name())
			break
		}
	}
	if extractedDir == "" {
		return fmt.Errorf("failed to find extracted nodejs dir")
	}
	if extractedDir != installDir {
		os.RemoveAll(installDir)
		if err := os.Rename(extractedDir, installDir); err != nil {
			return fmt.Errorf("failed to move extracted dir: %w", err)
		}
	}
	sep()

	// 修复 bin 目录下所有文件为可执行
	binDir := filepath.Join(installDir, "bin")
	filepath.Walk(binDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			os.Chmod(path, 0755)
		}
		return nil
	})

	title(fmt.Sprintf("Step 3/3: Node.js %s installed successfully!", version))
	fmt.Printf("[kver][nodejs] Installed at: %s\n", installDir)
	sep()
	installOk = true
	return nil
}

func extractTarGz(tarball, dest string) error {
	f, err := os.Open(tarball)
	if err != nil {
		return err
	}
	defer f.Close()
	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()
	tr := tar.NewReader(gzr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		outPath := filepath.Join(dest, hdr.Name)
		if hdr.FileInfo().IsDir() {
			os.MkdirAll(outPath, 0755)
			continue
		}
		os.MkdirAll(filepath.Dir(outPath), 0755)
		outFile, err := os.Create(outPath)
		if err != nil {
			return err
		}
		if _, err := io.Copy(outFile, tr); err != nil {
			outFile.Close()
			return err
		}
		outFile.Close()
	}
	return nil
}

// 递归拷贝目录，保留软链和文件内容
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		tgt := filepath.Join(dst, rel)
		// 先删除目标，防止软链被覆盖成空文件
		if _, err := os.Lstat(tgt); err == nil {
			os.Remove(tgt)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			link, err := os.Readlink(path)
			if err != nil {
				return err
			}
			return os.Symlink(link, tgt)
		}
		if info.IsDir() {
			return os.MkdirAll(tgt, info.Mode())
		}
		// 普通文件
		fSrc, err := os.Open(path)
		if err != nil {
			return err
		}
		defer fSrc.Close()
		fDst, err := os.OpenFile(tgt, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
		if err != nil {
			return err
		}
		defer fDst.Close()
		_, err = io.Copy(fDst, fSrc)
		return err
	})
}

func (n *NodejsPlugin) Uninstall(version string) error {
	home, _ := os.UserHomeDir()
	kverHome := filepath.Join(home, ".kver")
	installDir := filepath.Join(kverHome, "languages", "nodejs", version)
	envFile := filepath.Join(kverHome, "env.d", "nodejs.sh")
	os.Remove(envFile)
	if err := os.RemoveAll(installDir); err != nil {
		return fmt.Errorf("failed to remove nodejs version: %w", err)
	}
	fmt.Println("[kver] Node.js", version, "uninstalled.")
	return nil
}

func (n *NodejsPlugin) List() ([]string, error) {
	home, _ := os.UserHomeDir()
	base := filepath.Join(home, ".kver", "languages", "nodejs")
	entries, err := os.ReadDir(base)
	if err != nil {
		return nil, err
	}
	var versions []string
	for _, e := range entries {
		if e.IsDir() {
			versions = append(versions, e.Name())
		}
	}
	sort.Strings(versions)
	return versions, nil
}

func (n *NodejsPlugin) ListRemote() ([]string, error) {
	resp, err := http.Get("https://nodejs.org/dist/index.tab")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var versions []string
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "v") {
			fields := strings.Fields(line)
			if len(fields) > 0 {
				versions = append(versions, strings.TrimPrefix(fields[0], "v"))
			}
		}
	}
	return versions, nil
}

func (n *NodejsPlugin) Use(version string) error {
	home, _ := os.UserHomeDir()
	kverHome := filepath.Join(home, ".kver")
	installDir := filepath.Join(kverHome, "languages", "nodejs", version)
	binDir := filepath.Join(installDir, "bin")
	if _, err := os.Stat(installDir); os.IsNotExist(err) {
		return fmt.Errorf("nodejs version not installed: %s", version)
	}
	envDir := filepath.Join(kverHome, "env.d")
	os.MkdirAll(envDir, 0755)
	envFile := filepath.Join(envDir, "nodejs.sh")
	f, err := os.Create(envFile)
	if err != nil {
		return err
	}
	defer f.Close()
	fmt.Fprintf(f, "export NODEJS_HOME=\"%s\"\n", installDir)
	fmt.Fprintf(f, "export PATH=\"%s:$PATH\"\n", binDir)
	fmt.Println("[kver] Now using nodejs", version)
	return nil
}

func (n *NodejsPlugin) Global(version string) error {
	return n.Use(version)
}

func (n *NodejsPlugin) Local(version string, projectDir string) error {
	home, _ := os.UserHomeDir()
	kverHome := filepath.Join(home, ".kver")
	installDir := filepath.Join(kverHome, "languages", "nodejs", version)
	if _, err := os.Stat(installDir); os.IsNotExist(err) {
		return fmt.Errorf("nodejs version not installed: %s", version)
	}
	localFile := filepath.Join(projectDir, ".kver")
	f, err := os.OpenFile(localFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	fmt.Fprintf(f, "nodejs = %s\n", version)
	fmt.Println("[kver] Set local nodejs version to", version)
	return nil
}

func (n *NodejsPlugin) ActivateShell(version string) string {
	home, _ := os.UserHomeDir()
	installDir := filepath.Join(home, ".kver", "languages", "nodejs", version)
	return fmt.Sprintf("export NODEJS_HOME=\"%s\"\nexport PATH=\"$NODEJS_HOME/bin:$PATH\"\n", installDir)
}

func init() {
	plugin.Register("nodejs", &NodejsPlugin{})
}
