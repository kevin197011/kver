// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package goimpl

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"kver/internal/plugin"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
)

type GoPlugin struct{}

func (g *GoPlugin) Name() string { return "go" }

func (g *GoPlugin) Install(version string) error {
	home, _ := os.UserHomeDir()
	kverHome := filepath.Join(home, ".kver")
	installDir := filepath.Join(kverHome, "languages", "go", version)

	var installOk bool
	defer func() {
		if !installOk {
			os.RemoveAll(installDir)
		}
	}()

	tmpDir, err := os.MkdirTemp("", "kver-go-src-")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	title := func(s string) {
		fmt.Printf("\n\033[1;36m[kver][go] %s\033[0m\n", s)
	}
	sep := func() {
		fmt.Println("\033[1;34m----------------------------------------\033[0m")
	}

	title("Step 1/4: Download Go tarball")
	osStr := runtime.GOOS
	archStr := runtime.GOARCH
	goTarName := fmt.Sprintf("go%s.%s-%s.tar.gz", version, osStr, archStr)
	url := fmt.Sprintf("https://go.dev/dl/%s", goTarName)
	fmt.Printf("[kver][go] Downloading %s\n", url)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("download failed: %s", resp.Status)
	}
	tarball := filepath.Join(tmpDir, goTarName)
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

	title("Step 2/4: Extract Go tarball")
	if err := g.extractTarGz(tarball, tmpDir); err != nil {
		return err
	}
	sep()

	// 查找解压后的 go 目录
	srcDir := ""
	dirs, _ := os.ReadDir(tmpDir)
	for _, d := range dirs {
		if d.IsDir() && d.Name() == "go" {
			srcDir = filepath.Join(tmpDir, d.Name())
			break
		}
	}
	if srcDir == "" {
		return fmt.Errorf("failed to find extracted go dir")
	}

	title("Step 3/4: Move to install directory")
	os.RemoveAll(installDir)
	// 确保父目录存在
	os.MkdirAll(filepath.Dir(installDir), 0755)
	if err := os.Rename(srcDir, installDir); err != nil {
		return fmt.Errorf("failed to move go dir: %w", err)
	}
	sep()

	title(fmt.Sprintf("Step 4/4: Go %s installed successfully!", version))
	fmt.Printf("[kver][go] Installed at: %s\n", installDir)
	sep()
	installOk = true
	return nil
}

// extractTarGz 解压 tar.gz 包到目标目录
func (g *GoPlugin) extractTarGz(tarball, dest string) error {
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
		if hdr.Typeflag == tar.TypeDir {
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

func (g *GoPlugin) Uninstall(version string) error {
	home, _ := os.UserHomeDir()
	kverHome := filepath.Join(home, ".kver")
	installDir := filepath.Join(kverHome, "languages", "go", version)
	symlink := filepath.Join(kverHome, "versions", "go")

	if _, err := os.Stat(installDir); os.IsNotExist(err) {
		return fmt.Errorf("go version not installed: %s", version)
	}

	if link, err := os.Readlink(symlink); err == nil && link == installDir {
		os.Remove(symlink)
		os.Remove(filepath.Join(kverHome, "env.sh"))
	}

	if err := os.RemoveAll(installDir); err != nil {
		return fmt.Errorf("failed to remove go version: %w", err)
	}

	fmt.Println("[kver] Go", version, "uninstalled.")
	return nil
}

func (g *GoPlugin) List() ([]string, error) {
	home, _ := os.UserHomeDir()
	base := filepath.Join(home, ".kver", "languages", "go")
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

func (g *GoPlugin) ListRemote() ([]string, error) {
	resp, err := http.Get("https://go.dev/dl/")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var versions []string
	scanner := bufio.NewScanner(resp.Body)
	re := regexp.MustCompile(`go([0-9]+\.[0-9]+\.[0-9]+)\.`)
	for scanner.Scan() {
		line := scanner.Text()
		if m := re.FindStringSubmatch(line); m != nil {
			versions = append(versions, m[1])
		}
	}
	// 去重
	verMap := map[string]struct{}{}
	for _, v := range versions {
		verMap[v] = struct{}{}
	}
	versions = versions[:0]
	for v := range verMap {
		versions = append(versions, v)
	}
	sort.Strings(versions)
	return versions, nil
}

func (g *GoPlugin) Use(version string) error {
	home, _ := os.UserHomeDir()
	kverHome := filepath.Join(home, ".kver")
	installDir := filepath.Join(kverHome, "languages", "go", version)
	symlink := filepath.Join(kverHome, "versions", "go")
	os.MkdirAll(filepath.Dir(symlink), 0755)
	os.Remove(symlink)
	if err := os.Symlink(installDir, symlink); err != nil {
		return fmt.Errorf("failed to symlink: %w", err)
	}
	envsh := filepath.Join(kverHome, "env.sh")
	f, err := os.Create(envsh)
	if err != nil {
		return fmt.Errorf("failed to write env.sh: %w", err)
	}
	defer f.Close()
	f.WriteString(fmt.Sprintf("export GOROOT=\"%s\"\nexport PATH=\"$GOROOT/bin:$PATH\"\n", installDir))
	fmt.Printf("[kver] Now using go %s.\n", version)
	fmt.Printf("[kver] Please run: source %s\n", envsh)
	return nil
}

func (g *GoPlugin) Global(version string) error {
	return g.Use(version)
}

func (g *GoPlugin) Local(version string, projectDir string) error {
	kverFile := filepath.Join(projectDir, ".kver")
	f, err := os.Create(kverFile)
	if err != nil {
		return fmt.Errorf("failed to write .kver: %w", err)
	}
	defer f.Close()
	f.WriteString(fmt.Sprintf("go = %s\n", version))
	return g.Use(version)
}

func (g *GoPlugin) ActivateShell(version string) string {
	home, _ := os.UserHomeDir()
	installDir := filepath.Join(home, ".kver", "languages", "go", version)
	return fmt.Sprintf("export GOROOT=\"%s\"\nexport PATH=\"$GOROOT/bin:$PATH\"\n", installDir)
}

func init() {
	plugin.Register("go", &GoPlugin{})
}
