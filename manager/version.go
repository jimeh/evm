package manager

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

type Version struct {
	Version  string   `yaml:"version" json:"version"`
	Current  bool     `yaml:"current" json:"current"`
	Path     string   `yaml:"path" json:"path"`
	BinDir   string   `yaml:"bin_dir" json:"bin_dir"`
	Binaries []string `yaml:"binaries" json:"binaries"`
}

func (ver *Version) FindBin(name string) (string, error) {
	for _, b := range ver.Binaries {
		if filepath.Base(b) == name {
			return b, nil
		}
	}

	return "", fmt.Errorf(
		`%wExecutable "%s" not found in Emacs version %s`,
		ErrBinNotFound, name, ver.Version,
	)
}

func newVersion(
	ctx context.Context,
	conf *Config,
	version string,
) (*Version, error) {
	if version == "" {
		return nil, fmt.Errorf("%wversion cannot be empty", ErrVersion)
	}

	path := filepath.Join(conf.Paths.Versions, version)

	_, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf(
			"%wVersion %s is not available in %s",
			ErrVersionNotFound, version, conf.Paths.Versions,
		)
	}

	ver := &Version{
		Version: version,
		Path:    path,
		BinDir:  filepath.Join(path, "bin"),
		Current: version == conf.Current.Version,
	}

	entries, err := os.ReadDir(ver.BinDir)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}

	for _, entry := range entries {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		binPath := filepath.Join(ver.BinDir, entry.Name())

		f, err := os.Stat(binPath)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			return nil, err
		}

		// Ensure f is Regular file and executable.
		if f.Mode().IsRegular() && f.Mode().Perm()&0111 == 0111 {
			ver.Binaries = append(ver.Binaries, binPath)
		}
	}

	return ver, nil
}

func newVersions(ctx context.Context, conf *Config) ([]*Version, error) {
	results := []*Version{}

	entries, err := os.ReadDir(conf.Paths.Versions)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}

	for _, entry := range entries {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		if !entry.IsDir() {
			continue
		}

		ver, err := newVersion(ctx, conf, entry.Name())
		if err != nil {
			return nil, err
		}

		results = append(results, ver)
	}

	return results, nil
}
