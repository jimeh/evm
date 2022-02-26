package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

type emacsVersions []*emacsVersion

func newEmacsVersions(conf *config) (emacsVersions, error) {
	results := emacsVersions{}

	entries, err := os.ReadDir(conf.Path.Versions)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		vi, err := newEmacsVersion(conf, entry.Name())
		if err != nil {
			return nil, err
		}

		results = append(results, vi)
	}

	return results, nil
}

type emacsVersion struct {
	Version  string   `yaml:"version" json:"version"`
	Current  bool     `yaml:"current" json:"current"`
	Path     string   `yaml:"path" json:"path"`
	Bin      string   `yaml:"bin" json:"bin"`
	Binaries []string `yaml:"binaries" json:"binaries"`
}

func (ev *emacsVersion) BinPath(name string) (string, bool) {
	for _, b := range ev.Binaries {
		if filepath.Base(b) == name {
			return filepath.Join(ev.Bin, b), true
		}
	}

	return "", false
}

func newEmacsVersion(conf *config, version string) (*emacsVersion, error) {
	path := filepath.Join(conf.Path.Versions, version)

	_, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf(
			"Version %s is not available in %s",
			version, conf.Path.Versions,
		)
	}

	ev := &emacsVersion{
		Version: version,
		Path:    path,
		Bin:     filepath.Join(path, "bin"),
		Current: version == conf.Current.Version,
	}

	entries, err := os.ReadDir(ev.Bin)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}

	for _, entry := range entries {
		binPath := filepath.Join(ev.Bin, entry.Name())

		f, err := os.Stat(binPath)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			return nil, err
		}

		relPath, err := filepath.Rel(ev.Bin, binPath)
		if err != nil {
			return nil, err
		}

		// Regular and executable file.
		if f.Mode().IsRegular() && f.Mode().Perm()&0111 == 0111 {
			ev.Binaries = append(ev.Binaries, relPath)
		}
	}

	return ev, nil
}
