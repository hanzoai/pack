package mise

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

const (
	InstallDir     = "/tmp/railpack/mise"
	TestInstallDir = "/tmp/railpack/mise-test"
)

type Mise struct {
	binaryPath string
	cacheDir   string
}

func New(cacheDir string) (*Mise, error) {
	binaryPath, err := ensureInstalled(cacheDir)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure mise is installed: %w", err)
	}

	return &Mise{
		binaryPath: binaryPath,
		cacheDir:   cacheDir,
	}, nil
}

// GetLatestVersion gets the latest version of a package matching the version constraint
func (m *Mise) GetLatestVersion(pkg, version string) (string, error) {
	query := fmt.Sprintf("%s@%s", pkg, strings.TrimSpace(version))
	output, err := m.runCmd("latest", "--verbose", query)
	if err != nil {
		return "", err
	}

	latestVersion := strings.TrimSpace(output)
	if latestVersion == "" {
		return "", fmt.Errorf("failed to get latest version for %s", query)
	}

	return latestVersion, nil
}

// runCmd runs a mise command with the given arguments
func (m *Mise) runCmd(args ...string) (string, error) {
	// Use persistent directories for mise data and cache
	cacheDir := filepath.Join(m.cacheDir, "cache")
	dataDir := filepath.Join(m.cacheDir, "data")

	// mu, err := filemutex.New("/tmp/foo.lock")
	// if err != nil {
	// 	return "", fmt.Errorf("failed to create mutex: %w", err)
	// }

	// mu.Lock() // Will block until lock can be acquired
	// defer mu.Unlock()

	cmd := exec.Command(m.binaryPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// We want to shell out to the git CLI here
	// Without it, I was noticing races when multiple processes tried to check the version of the same package in parallel
	// Sometimes a checkout operation would fail.
	// I am testing out forcing usage of the git CLI to see if it helps
	// Source: https://github.com/jdx/mise/blob/main/src/git.rs#L124
	// Config: https://github.com/jdx/mise/blob/main/settings.toml#L369
	// cmd.Env = append(cmd.Env, "MISE_LIBGIT2=false")
	// cmd.Env = append(cmd.Env, "MISE_GIX=false")
	// cmd.Env = append(cmd.Env, "RUST_BACKTRACE=full")

	cmd.Env = append(cmd.Env,
		fmt.Sprintf("MISE_CACHE_DIR=%s", cacheDir),
		fmt.Sprintf("MISE_DATA_DIR=%s", dataDir),
		fmt.Sprintf("PATH=%s", os.Getenv("PATH")),
	)

	// cmd.Env = append(cmd.Env, os.Environ()...)

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to run mise command '%s': %w\nstdout: %s\nstderr: %s",
			strings.Join(append([]string{m.binaryPath}, args...), " "),
			err,
			stdout.String(),
			stderr.String())
	}

	return stdout.String(), nil
}

// MisePackage represents a single mise package configuration
type MisePackage struct {
	Version string `toml:"version"`
}

// MiseConfig represents the overall mise configuration
type MiseConfig struct {
	Tools map[string]MisePackage `toml:"tools"`
}

func GenerateMiseToml(packages map[string]string) (string, error) {
	config := MiseConfig{
		Tools: make(map[string]MisePackage),
	}

	for name, version := range packages {
		config.Tools[name] = MisePackage{Version: version}
	}

	buf := bytes.NewBuffer(nil)
	if err := toml.NewEncoder(buf).Encode(config); err != nil {
		return "", err
	}

	return buf.String(), nil
}
