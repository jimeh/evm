package main

import (
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

const currentFileName = "current"

var (
	configUnmashaled bool
	cachedConfig     = &config{}
)

type config struct {
	Mode    string        `yaml:"mode" json:"mode"`
	Current currentConfig `yaml:"current" json:"current"`
	Path    pathsConfig   `yaml:"path" json:"path"`
}

type currentConfig struct {
	Version string `yaml:"version" json:"version"`
	SetBy   string `yaml:"set_by" json:"set_by"`
}

type pathsConfig struct {
	Binary   string `yaml:"binary" json:"binary"`
	Root     string `yaml:"root" json:"root"`
	Shims    string `yaml:"shims" json:"shims"`
	Sources  string `yaml:"sources" json:"sources"`
	Versions string `yaml:"versions" json:"versions"`
}

func getConfig() (*config, error) {
	if configUnmashaled {
		cc, err := getCurrentVersion(cachedConfig)
		if err != nil {
			return nil, err
		}

		cachedConfig.Current = *cc

		return cachedConfig, nil
	}

	conf := &config{}
	err := viper.Unmarshal(conf)
	if err != nil {
		return nil, err
	}

	if conf.Mode != "user" && conf.Mode != "system" {
		return nil, errors.New(`When set EVM_MODE must be "user" or "system"`)
	}

	conf.Path.Binary, err = os.Executable()
	if err != nil {
		return nil, err
	}

	var homePrefix string
	switch {
	case strings.HasPrefix(conf.Path.Root, "$HOME") ||
		strings.HasPrefix(conf.Path.Root, "$home"):
		homePrefix = conf.Path.Root[0:5]
	case strings.HasPrefix(conf.Path.Root, "~"):
		homePrefix = conf.Path.Root[0:1]
	}

	if homePrefix != "" {
		var home string
		home, err = os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		conf.Path.Root = filepath.Join(
			home, strings.TrimPrefix(conf.Path.Root, homePrefix))
	}

	if strings.HasPrefix(conf.Path.Shims, "$EVM_ROOT") {
		conf.Path.Shims = filepath.Join(
			conf.Path.Root, conf.Path.Shims[9:],
		)
	} else if !path.IsAbs(conf.Path.Shims) {
		conf.Path.Shims = filepath.Join(
			conf.Path.Root, conf.Path.Shims,
		)
	}

	if strings.HasPrefix(conf.Path.Sources, "$EVM_ROOT") {
		conf.Path.Sources = filepath.Join(
			conf.Path.Root, conf.Path.Sources[9:],
		)
	} else if !path.IsAbs(conf.Path.Sources) {
		conf.Path.Sources = filepath.Join(
			conf.Path.Root, conf.Path.Sources,
		)
	}

	if strings.HasPrefix(conf.Path.Versions, "$EVM_ROOT") {
		conf.Path.Versions = filepath.Join(
			conf.Path.Root, conf.Path.Versions[9:],
		)
	} else if !path.IsAbs(conf.Path.Versions) {
		conf.Path.Versions = filepath.Join(
			conf.Path.Root, conf.Path.Versions,
		)
	}

	cc, err := getCurrentVersion(conf)
	if err != nil {
		return nil, err
	}

	conf.Current = *cc
	cachedConfig = conf

	return cachedConfig, nil
}

func getCurrentVersion(conf *config) (*currentConfig, error) {
	cc := &currentConfig{}

	if v := os.Getenv("EVM_VERSION"); v != "" {
		cc.Version = strings.TrimSpace(v)
		cc.SetBy = "EVM_VERSION environment variable"

		return cc, nil
	}

	currentFile := filepath.Join(conf.Path.Root, currentFileName)
	b, err := os.ReadFile(currentFile)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return cc, nil
		}
		return nil, err
	}

	if len(b) > 0 {
		cc.Version = strings.TrimSpace(string(b))
		cc.SetBy = currentFile
	}

	return cc, nil
}
