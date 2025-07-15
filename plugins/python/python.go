// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package python

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"kver/internal/plugin"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

type PythonPlugin struct{}

func (p *PythonPlugin) Name() string { return "python" }

func (p *PythonPlugin) Install(version string) error {
	home, _ := os.UserHomeDir()
	kverHome := filepath.Join(home, ".kver")
	installDir := filepath.Join(kverHome, "languages", "python", version)

	var installOk bool
	defer func() {
		if !installOk {
			os.RemoveAll(installDir)
		}
	}()

	tmpDir, err := os.MkdirTemp("", "kver-python-src-")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	title := func(s string) {
		fmt.Printf("\n\033[1;36m[kver][python] %s\033[0m\n", s)
	}
	sep := func() {
		fmt.Println("\033[1;34m----------------------------------------\033[0m")
	}

	title("Step 1/5: Download Python tarball")
	url := fmt.Sprintf("https://www.python.org/ftp/python/%s/Python-%s.tgz", version, version)
	fmt.Printf("[kver][python] Downloading %s\n", url)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("download failed: %s", resp.Status)
	}
	tarball := filepath.Join(tmpDir, fmt.Sprintf("Python-%s.tgz", version))
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

	title("Step 2/5: Extract Python source")
	if err := p.extractTarGz(tarball, tmpDir); err != nil {
		return err
	}
	sep()

	// 查找解压后的源码目录
	srcDir := ""
	dirs, _ := os.ReadDir(tmpDir)
	for _, d := range dirs {
		if d.IsDir() && strings.HasPrefix(d.Name(), "Python-") {
			srcDir = filepath.Join(tmpDir, d.Name())
			break
		}
	}
	if srcDir == "" {
		return fmt.Errorf("failed to find extracted python source dir")
	}

	// 修正 configure 权限
	configurePath := filepath.Join(srcDir, "configure")
	if err := os.Chmod(configurePath, 0755); err != nil {
		return fmt.Errorf("failed to chmod configure: %w", err)
	}
	if err := fixExecPerms(srcDir); err != nil {
		return fmt.Errorf("failed to fix exec perms: %w", err)
	}

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

	// 自动补软链
	binDir := filepath.Join(installDir, "bin")
	parts := strings.Split(version, ".")
	pyMajor := parts[0]
	pyMinor := parts[1]
	pyPatch := ""
	if len(parts) > 2 {
		pyPatch = parts[2]
	}
	pyVer := pyMajor + "." + pyMinor
	if pyPatch != "" {
		pyVer = pyVer + "." + pyPatch
	}
	pythonExe := "python" + pyMajor + "." + pyMinor
	if pyPatch != "" {
		pythonExe = "python" + pyMajor + "." + pyMinor + "." + pyPatch
	}
	pipExe := "pip" + pyMajor + "." + pyMinor
	if pyPatch != "" {
		pipExe = "pip" + pyMajor + "." + pyMinor + "." + pyPatch
	}

	// python3 -> python3.x[.x]
	if _, err := os.Stat(filepath.Join(binDir, "python3")); os.IsNotExist(err) {
		if _, err2 := os.Stat(filepath.Join(binDir, pythonExe)); err2 == nil {
			os.Symlink(pythonExe, filepath.Join(binDir, "python3"))
		}
	}
	// python -> python3
	if _, err := os.Stat(filepath.Join(binDir, "python")); os.IsNotExist(err) {
		if _, err2 := os.Stat(filepath.Join(binDir, "python3")); err2 == nil {
			os.Symlink("python3", filepath.Join(binDir, "python"))
		}
	}
	// pip3 -> pip3.x[.x]
	if _, err := os.Stat(filepath.Join(binDir, "pip3")); os.IsNotExist(err) {
		if _, err2 := os.Stat(filepath.Join(binDir, pipExe)); err2 == nil {
			os.Symlink(pipExe, filepath.Join(binDir, "pip3"))
		}
	}
	// pip -> pip3
	if _, err := os.Stat(filepath.Join(binDir, "pip")); os.IsNotExist(err) {
		if _, err2 := os.Stat(filepath.Join(binDir, "pip3")); err2 == nil {
			os.Symlink("pip3", filepath.Join(binDir, "pip"))
		}
	}

	title(fmt.Sprintf("Python %s installed successfully!", version))
	fmt.Printf("[kver][python] Installed at: %s\n", installDir)
	sep()
	installOk = true
	return nil
}

// extractTarGz 解压 tar.gz 包到目标目录
func (p *PythonPlugin) extractTarGz(tarball, dest string) error {
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

func (p *PythonPlugin) Uninstall(version string) error {
	home, _ := os.UserHomeDir()
	kverHome := filepath.Join(home, ".kver")
	installDir := filepath.Join(kverHome, "languages", "python", version)
	envFile := filepath.Join(kverHome, "env.d", "python.sh")
	os.Remove(envFile)
	if err := os.RemoveAll(installDir); err != nil {
		return fmt.Errorf("failed to remove python version: %w", err)
	}
	fmt.Println("[kver] Python", version, "uninstalled.")
	return nil
}

func (p *PythonPlugin) List() ([]string, error) {
	home, _ := os.UserHomeDir()
	base := filepath.Join(home, ".kver", "languages", "python")
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

func (p *PythonPlugin) ListRemote() ([]string, error) {
	// 官方页面 https://www.python.org/ftp/python/ 有目录索引
	resp, err := http.Get("https://www.python.org/ftp/python/")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var versions []string
	scanner := bufio.NewScanner(resp.Body)
	re := regexp.MustCompile(`>([0-9]+\.[0-9]+\.[0-9]+)/<`)
	for scanner.Scan() {
		line := scanner.Text()
		if m := re.FindStringSubmatch(line); m != nil {
			versions = append(versions, m[1])
		}
	}
	sort.Strings(versions)
	return versions, nil
}

func (p *PythonPlugin) Use(version string) error {
	home, _ := os.UserHomeDir()
	kverHome := filepath.Join(home, ".kver")
	installDir := filepath.Join(kverHome, "languages", "python", version)
	binDir := filepath.Join(installDir, "bin")
	if _, err := os.Stat(installDir); os.IsNotExist(err) {
		return fmt.Errorf("python version not installed: %s", version)
	}
	envDir := filepath.Join(kverHome, "env.d")
	os.MkdirAll(envDir, 0755)
	envFile := filepath.Join(envDir, "python.sh")
	f, err := os.Create(envFile)
	if err != nil {
		return err
	}
	defer f.Close()
	fmt.Fprintf(f, "export PYTHON_HOME=\"%s\"\n", installDir)
	fmt.Fprintf(f, "export PATH=\"%s:$PATH\"\n", binDir)
	fmt.Println("[kver] Now using python", version)
	return nil
}

func (p *PythonPlugin) Global(version string) error {
	return p.Use(version)
}

func (p *PythonPlugin) Local(version string, projectDir string) error {
	home, _ := os.UserHomeDir()
	kverHome := filepath.Join(home, ".kver")
	installDir := filepath.Join(kverHome, "languages", "python", version)
	if _, err := os.Stat(installDir); os.IsNotExist(err) {
		return fmt.Errorf("python version not installed: %s", version)
	}
	localFile := filepath.Join(projectDir, ".kver")
	f, err := os.OpenFile(localFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	fmt.Fprintf(f, "python = %s\n", version)
	fmt.Println("[kver] Set local python version to", version)
	return nil
}

func (p *PythonPlugin) ActivateShell(version string) string {
	home, _ := os.UserHomeDir()
	installDir := filepath.Join(home, ".kver", "languages", "python", version)
	return fmt.Sprintf("export PYTHON_HOME=\"%s\"\nexport PATH=\"$PYTHON_HOME/bin:$PATH\"\n", installDir)
}

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
	plugin.Register("python", &PythonPlugin{})
}
