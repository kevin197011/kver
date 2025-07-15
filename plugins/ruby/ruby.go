// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package ruby

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"io/fs"
	"kver/internal/plugin"
)

type RubyPlugin struct{}

func (r *RubyPlugin) Name() string { return "ruby" }

func (r *RubyPlugin) Install(version string) error {
	home, _ := os.UserHomeDir()
	kverHome := filepath.Join(home, ".kver")
	installDir := filepath.Join(kverHome, "languages", "ruby", version)

	var installOk bool
	defer func() {
		if !installOk {
			os.RemoveAll(installDir)
		}
	}()

	tmpDir, err := os.MkdirTemp("", "kver-ruby-src-")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	title := func(s string) {
		fmt.Printf("\n\033[1;36m[kver][ruby] %s\033[0m\n", s)
	}
	sep := func() {
		fmt.Println("\033[1;34m----------------------------------------\033[0m")
	}

	title("Step 1/5: Download Ruby tarball")
	majorMinor := version[:strings.LastIndex(version, ".")]
	url := fmt.Sprintf("https://cache.ruby-lang.org/pub/ruby/%s/ruby-%s.tar.gz", majorMinor, version)
	fmt.Printf("[kver][ruby] Downloading %s\n", url)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("download failed: %s", resp.Status)
	}
	tarball := filepath.Join(tmpDir, fmt.Sprintf("ruby-%s.tar.gz", version))
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

	title("Step 2/5: Extract Ruby source")
	if err := r.extractTarGz(tarball, tmpDir); err != nil {
		return err
	}
	sep()

	// 查找解压后的源码目录
	srcDir := ""
	dirs, _ := os.ReadDir(tmpDir)
	for _, d := range dirs {
		if d.IsDir() && strings.HasPrefix(d.Name(), "ruby-") {
			srcDir = filepath.Join(tmpDir, d.Name())
			break
		}
	}
	if srcDir == "" {
		return fmt.Errorf("failed to find extracted ruby source dir")
	}

	// 修正 configure 权限
	configurePath := filepath.Join(srcDir, "configure")
	if err := os.Chmod(configurePath, 0755); err != nil {
		return fmt.Errorf("failed to chmod configure: %w", err)
	}
	// 递归修正源码目录下所有脚本和工具的可执行权限
	if err := fixExecPerms(srcDir); err != nil {
		return fmt.Errorf("failed to fix exec perms: %w", err)
	}

	// 只有到这里才创建 installDir（如果 configure/make install 没有自动创建）
	// 但 Ruby 的 make install 会自动创建，不需要提前创建

	title("Step 3/5: Configure build")
	cmdConf := exec.Command("./configure", "--prefix="+installDir)
	cmdConf.Dir = srcDir
	cmdConf.Stdout = os.Stdout
	cmdConf.Stderr = os.Stderr
	if err := cmdConf.Run(); err != nil {
		return fmt.Errorf("configure failed: %w", err)
	}
	sep()

	title("Step 4/5: Compile (make -jN)")
	cpuCount := runtime.NumCPU()
	makeArgs := []string{"-j" + strconv.Itoa(cpuCount)}
	cmdMake := exec.Command("make", makeArgs...)
	cmdMake.Dir = srcDir
	cmdMake.Stdout = os.Stdout
	cmdMake.Stderr = os.Stderr
	if err := cmdMake.Run(); err != nil {
		return fmt.Errorf("make failed: %w", err)
	}
	sep()

	title("Step 5/5: Install to target directory")
	cmdInstall := exec.Command("make", append(makeArgs, "install")...)
	cmdInstall.Dir = srcDir
	cmdInstall.Stdout = os.Stdout
	cmdInstall.Stderr = os.Stderr
	if err := cmdInstall.Run(); err != nil {
		return fmt.Errorf("make install failed: %w", err)
	}
	sep()

	title(fmt.Sprintf("Ruby %s installed successfully!", version))
	fmt.Printf("[kver][ruby] Installed at: %s\n", installDir)
	sep()
	installOk = true
	return nil
}

// extractTarGz 解压 tar.gz 包到目标目录
func (r *RubyPlugin) extractTarGz(tarball, dest string) error {
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

// buildAndInstallRuby 编译并安装 Ruby 到指定目录
func (r *RubyPlugin) buildAndInstallRuby(srcDir, prefix string) error {
	configure := exec.Command("./configure", "--prefix="+prefix)
	configure.Dir = srcDir
	configure.Stdout = os.Stdout
	configure.Stderr = os.Stderr
	if err := configure.Run(); err != nil {
		return fmt.Errorf("configure failed: %w", err)
	}

	cpu := runtime.NumCPU()
	make := exec.Command("make", fmt.Sprintf("-j%d", cpu))
	make.Dir = srcDir
	make.Stdout = os.Stdout
	make.Stderr = os.Stderr
	if err := make.Run(); err != nil {
		return fmt.Errorf("make failed: %w", err)
	}

	makeInstall := exec.Command("make", "install")
	makeInstall.Dir = srcDir
	makeInstall.Stdout = os.Stdout
	makeInstall.Stderr = os.Stderr
	if err := makeInstall.Run(); err != nil {
		return fmt.Errorf("make install failed: %w", err)
	}
	return nil
}

func (r *RubyPlugin) Uninstall(version string) error {
	home, _ := os.UserHomeDir()
	kverHome := filepath.Join(home, ".kver")
	installDir := filepath.Join(kverHome, "languages", "ruby", version)
	symlink := filepath.Join(kverHome, "versions", "ruby")

	if _, err := os.Stat(installDir); os.IsNotExist(err) {
		return fmt.Errorf("ruby version not installed: %s", version)
	}

	// 如果软链指向该版本，移除软链和 env.sh
	if link, err := os.Readlink(symlink); err == nil && link == installDir {
		os.Remove(symlink)
		os.Remove(filepath.Join(kverHome, "env.sh"))
	}

	if err := os.RemoveAll(installDir); err != nil {
		return fmt.Errorf("failed to remove ruby version: %w", err)
	}

	fmt.Println("[kver] Ruby", version, "uninstalled.")
	return nil
}

func (r *RubyPlugin) List() ([]string, error) {
	home, _ := os.UserHomeDir()
	base := filepath.Join(home, ".kver", "languages", "ruby")
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
	return versions, nil
}

func (r *RubyPlugin) ListRemote() ([]string, error) {
	resp, err := http.Get("https://cache.ruby-lang.org/pub/ruby/index.txt")
	if err != nil {
		return []string{"3.3.0", "3.2.2", "3.1.4", "2.7.8"}, nil
	}
	defer resp.Body.Close()

	re := regexp.MustCompile(`ruby-([0-9]+\.[0-9]+\.[0-9]+)\.tar\.gz`)
	verMap := make(map[string]struct{})
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if m := re.FindStringSubmatch(line); m != nil {
			verMap[m[1]] = struct{}{}
		}
	}
	if err := scanner.Err(); err != nil {
		return []string{"3.3.0", "3.2.2", "3.1.4", "2.7.8"}, nil
	}
	uniq := make([]string, 0, len(verMap))
	for v := range verMap {
		uniq = append(uniq, v)
	}
	sort.Strings(uniq)
	if len(uniq) == 0 {
		return []string{"3.3.0", "3.2.2", "3.1.4", "2.7.8"}, nil
	}
	return uniq, nil
}

func (r *RubyPlugin) Use(version string) error {
	home, _ := os.UserHomeDir()
	kverHome := filepath.Join(home, ".kver")
	installDir := filepath.Join(kverHome, "languages", "ruby", version)
	symlink := filepath.Join(kverHome, "versions", "ruby")

	if _, err := os.Stat(installDir); os.IsNotExist(err) {
		return fmt.Errorf("ruby version not installed: %s", version)
	}

	os.MkdirAll(filepath.Dir(symlink), 0755)
	os.Remove(symlink)
	if err := os.Symlink(installDir, symlink); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	envFile := filepath.Join(kverHome, "env.sh")
	f, err := os.Create(envFile)
	if err != nil {
		return fmt.Errorf("failed to write env.sh: %w", err)
	}
	defer f.Close()
	f.WriteString(fmt.Sprintf("export RUBY_HOME=%s\nexport PATH=\"$RUBY_HOME/bin:$PATH\"\n", symlink))

	fmt.Println("[kver] Now using Ruby", version)
	fmt.Println("[kver] Please run: source ~/.kver/env.sh")
	return nil
}

func (r *RubyPlugin) Global(version string) error {
	return r.Use(version)
}

func (r *RubyPlugin) Local(version string, projectDir string) error {
	kverFile := filepath.Join(projectDir, ".kver")
	f, err := os.Create(kverFile)
	if err != nil {
		return fmt.Errorf("failed to write .kver: %w", err)
	}
	defer f.Close()
	f.WriteString(fmt.Sprintf("ruby = %s\n", version))
	return r.Use(version)
}

// ActivateShell 输出 shell 片段用于激活指定 Ruby 版本
func (r *RubyPlugin) ActivateShell(version string) string {
	home, _ := os.UserHomeDir()
	installDir := filepath.Join(home, ".kver", "languages", "ruby", version)
	return fmt.Sprintf("export RUBY_HOME=\"%s\"\nexport PATH=\"$RUBY_HOME/bin:$PATH\"\n", installDir)
}

// 修正源码目录下所有可执行文件权限
func fixExecPerms(root string) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		base := filepath.Base(path)
		if strings.HasSuffix(path, ".sh") ||
			strings.HasPrefix(base, "ifchange") ||
			strings.HasPrefix(base, "configure") ||
			strings.Contains(path, "/tool/") {
			return os.Chmod(path, 0755)
		}
		return nil
	})
}

func init() {
	plugin.Register("ruby", &RubyPlugin{})
}
