package manager

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/sethvargo/go-envconfig"
	"gopkg.in/yaml.v3"
)

var ErrConfig = fmt.Errorf("%w", Err)

type Mode string

const (
	User   Mode = "user"
	System Mode = "system"
)

type ConfigFile struct {
	Paths ConfigFilePaths `yaml:"paths" json:"paths"`
}

type ConfigFilePaths struct {
	Shims    string `yaml:"shims" json:"shims" env:"EVM_SHIMS,overwrite"`
	Sources  string `yaml:"sources" json:"sources"  env:"EVM_SOURCES,overwrite"`
	Versions string `yaml:"versions" json:"versions" env:"EVM_VERSIONS,overwrite"`
}

type Config struct {
	Mode    Mode          `yaml:"mode" json:"mode"`
	Current CurrentConfig `yaml:"current" json:"current"`
	Paths   PathsConfig   `yaml:"paths" json:"paths"`
}

type CurrentConfig struct {
	Version string `yaml:"version" json:"version"`
	SetBy   string `yaml:"set_by,omitempty" json:"set_by,omitempty"`
}

type PathsConfig struct {
	Binary   string `yaml:"binary" json:"binary"`
	Root     string `yaml:"root" json:"root"`
	Shims    string `yaml:"shims" json:"shims"`
	Sources  string `yaml:"sources" json:"sources"`
	Versions string `yaml:"versions" json:"versions"`
}

func NewConfig() (*Config, error) {
	mode := Mode(os.Getenv("EVM_MODE"))
	if mode != System {
		mode = User
	}

	defaultRoot := filepath.Join(string(os.PathSeparator), "opt", "evm")
	if mode == User {
		defaultRoot = filepath.Join("$HOME", ".evm")
	}

	if v := os.Getenv("EVM_ROOT"); v != "" {
		defaultRoot = v
	}

	conf := &Config{
		Mode: mode,
		Paths: PathsConfig{
			Root:     defaultRoot,
			Shims:    "$EVM_ROOT/shims",
			Sources:  "$EVM_ROOT/sources",
			Versions: "$EVM_ROOT/versions",
		},
	}

	var err error
	conf.Paths.Root, err = conf.normalizePath(conf.Paths.Root)
	if err != nil {
		return nil, err
	}

	err = conf.load()
	if err != nil {
		return nil, err
	}

	conf.Paths.Shims, err = conf.normalizePath(conf.Paths.Shims)
	if err != nil {
		return nil, err
	}
	conf.Paths.Sources, err = conf.normalizePath(conf.Paths.Sources)
	if err != nil {
		return nil, err
	}
	conf.Paths.Versions, err = conf.normalizePath(conf.Paths.Versions)
	if err != nil {
		return nil, err
	}

	conf.Paths.Binary, err = os.Executable()
	if err != nil {
		return nil, err
	}

	err = conf.PopulateCurrent()
	if err != nil {
		return nil, err
	}

	return conf, nil
}

const currentFileName = "current"

func (conf *Config) PopulateCurrent() error {
	if v := os.Getenv("EVM_VERSION"); v != "" {
		conf.Current.Version = strings.TrimSpace(v)
		conf.Current.SetBy = "EVM_VERSION environment variable"

		return nil
	}

	currentFile := filepath.Join(conf.Paths.Root, currentFileName)
	b, err := os.ReadFile(currentFile)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return err
	}

	if len(b) > 0 {
		conf.Current.Version = strings.TrimSpace(string(b))
		conf.Current.SetBy = currentFile
	}

	return nil
}

var configFileNames = []string{
	"config.yaml",
	"config.yml",
	"config.json",
	"evm.yaml",
	"evm.yml",
	"evm.json",
}

func (c *Config) load() error {
	var path string
	for _, name := range configFileNames {
		f := filepath.Join(c.Paths.Root, name)

		_, err := os.Stat(f)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			return err
		}

		path = f
		break
	}

	cf := &ConfigFile{}
	if path != "" {
		var err error
		cf, err = c.loadConfigFile(path)
		if err != nil {
			return err
		}
	}

	err := envconfig.Process(context.Background(), cf)
	if err != nil {
		return err
	}

	if cf.Paths.Shims != "" {
		c.Paths.Shims = cf.Paths.Shims
	}
	if cf.Paths.Sources != "" {
		c.Paths.Sources = cf.Paths.Sources
	}
	if cf.Paths.Versions != "" {
		c.Paths.Versions = cf.Paths.Versions
	}

	return nil
}

func (c *Config) loadConfigFile(path string) (*ConfigFile, error) {
	if path == "" {
		return nil, nil
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cf := &ConfigFile{}

	buf := bytes.NewBuffer(content)
	switch filepath.Ext(path) {
	case ".yaml", ".yml":
		dec := yaml.NewDecoder(buf)
		dec.KnownFields(true)
		err = dec.Decode(cf)
	case ".json":
		dec := json.NewDecoder(buf)
		dec.DisallowUnknownFields()
		err = dec.Decode(cf)
	default:
		return nil, fmt.Errorf(
			`%w"%s" does not have a ".yaml", ".yml", `+
				`or ".json" file extension`,
			ErrConfig, path,
		)
	}
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}

	return cf, nil
}

func (c *Config) normalizePath(path string) (string, error) {
	path = strings.TrimSpace(path)

	var homePrefix string
	switch {
	case strings.HasPrefix(path, "$HOME") ||
		strings.HasPrefix(path, "$home"):
		homePrefix = path[0:5]
	case strings.HasPrefix(path, "~"):
		homePrefix = path[0:1]
	}

	if homePrefix != "" {
		if c.Mode == System {
			return "", fmt.Errorf(
				`%wEVM_MODE is set to "%s" which prohibits `+
					`using "$HOME" or "~" in EVM_ROOT`,
				ErrConfig, string(System),
			)
		}

		var home string
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(
			home, strings.TrimPrefix(path, homePrefix))
	}

	if c.Paths.Root == "" {
		return path, nil
	}

	if strings.HasPrefix(path, "$EVM_ROOT") {
		path = filepath.Join(c.Paths.Root, path[9:])
	} else if !filepath.IsAbs(path) {
		path = filepath.Join(c.Paths.Root, path)
	}

	return path, nil
}
