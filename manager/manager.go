package manager

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/rs/zerolog/log"
)

var (
	Err                 = errors.New("")
	ErrVersion          = fmt.Errorf("%w", Err)
	ErrNoCurrentVersion = fmt.Errorf("%w", ErrVersion)
	ErrVersionNotFound  = fmt.Errorf("%w", ErrVersion)
	ErrBinNotFound      = fmt.Errorf("%w", ErrVersion)
)

type Manager struct {
	Config *Config
}

func New(config *Config) (*Manager, error) {
	if config == nil {
		var err error
		config, err = NewConfig()
		if err != nil {
			return nil, err
		}
	}

	return &Manager{Config: config}, nil
}

func (m *Manager) CurrentVersion() string {
	return m.Config.Current.Version
}

func (m *Manager) CurrentSetBy() string {
	return m.Config.Current.SetBy
}

func (m *Manager) List(ctx context.Context) ([]*Version, error) {
	return newVersions(ctx, m.Config)
}

func (m *Manager) Get(ctx context.Context, version string) (*Version, error) {
	return newVersion(ctx, m.Config, version)
}

func (m *Manager) Use(ctx context.Context, version string) error {
	log.Debug().Str("version", version).Msg("use version")

	ver, err := m.Get(ctx, version)
	if err != nil {
		return err
	}

	currentFile := filepath.Join(m.Config.Paths.Root, currentFileName)

	log.Debug().
		Str("path", currentFile).
		Str("content", ver.Version).
		Msg("updating current file")

	err = os.WriteFile(currentFile, []byte(ver.Version), 0o644)
	if err != nil {
		return err
	}

	err = m.rehashVersions(ctx, false, []*Version{ver})
	if err != nil {
		return err
	}

	return nil
}

func (m *Manager) RehashAll(ctx context.Context) error {
	versions, err := m.List(ctx)
	if err != nil {
		return err
	}

	return m.rehashVersions(ctx, true, versions)
}

func (m *Manager) RehashVersions(
	ctx context.Context,
	versions []string,
) error {
	var vers []*Version
	for _, s := range versions {
		v, err := m.Get(ctx, s)
		if err != nil {
			return err
		}

		vers = append(vers, v)
	}

	return m.rehashVersions(ctx, false, vers)
}

func (m *Manager) rehashVersions(
	ctx context.Context,
	tidy bool,
	versions []*Version,
) error {
	if log.Debug().Enabled() {
		var vers []string
		for _, v := range versions {
			vers = append(vers, v.Version)
		}
		log.Debug().Strs("versions", vers).Msg("reshashing versions")
	}

	programs := map[string]bool{}
	for _, ver := range versions {
		for _, bin := range ver.Binaries {
			base := filepath.Base(bin)
			programs[base] = true
		}
	}

	log.Debug().
		Str("path", m.Config.Paths.Shims).
		Msg("ensure shims directory exists")
	err := os.MkdirAll(m.Config.Paths.Shims, 0o755)
	if err != nil {
		return err
	}

	shims, err := m.ListShims(ctx)
	if err != nil {
		return err
	}

	shimMap := map[string]bool{}
	for _, s := range shims {
		base := filepath.Base(s)
		shimMap[base] = true
	}

	shim, err := m.shim()
	if err != nil {
		return err
	}

	for name := range programs {
		shimFile := filepath.Join(m.Config.Paths.Shims, name)

		log.Debug().Str("path", shimFile).Msg("writing shim")
		err = os.WriteFile(shimFile, shim, 0o755)
		if err != nil {
			return err
		}

		var f fs.FileInfo
		f, err = os.Stat(shimFile)
		if err != nil {
			return err
		}

		if f.Mode().Perm() != 0o755 {
			err = os.Chmod(shimFile, 0o755)
			if err != nil {
				return err
			}
		}

		delete(shimMap, name)
	}

	if tidy && len(shimMap) > 0 {
		log.Debug().Msg("tidying shims")
		for name := range shimMap {
			shimFile := filepath.Join(m.Config.Paths.Shims, name)
			log.Debug().Str("path", shimFile).Msg("removing shim")
			err := os.Remove(shimFile)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

var shimTemplate = template.Must(template.New("other").Parse(
	`#!/usr/bin/env bash
set -e
[ -n "$EVM_DEBUG" ] && set -x

program="${0##*/}"
export EVM_ROOT="{{.Paths.Root}}"
exec "{{.Paths.Binary}}" exec "$program" "$@"
`))

func (m *Manager) shim() ([]byte, error) {
	var buf bytes.Buffer
	err := shimTemplate.Execute(&buf, m.Config)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (m *Manager) ListShims(ctx context.Context) ([]string, error) {
	log.Debug().Str("path", m.Config.Paths.Shims).Msg("reading shims")

	entries, err := os.ReadDir(m.Config.Paths.Shims)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return []string{}, nil
		}
		return nil, err
	}

	r := []string{}

	for _, entry := range entries {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		shimPath := filepath.Join(m.Config.Paths.Shims, entry.Name())
		f, err := os.Stat(shimPath)
		if err != nil {
			return nil, err
		}

		if f.Mode().IsRegular() {
			r = append(r, shimPath)
		}
	}

	return r, nil
}

func (m *Manager) Exec(
	ctx context.Context,
	program string,
	args []string,
) error {
	current := m.CurrentVersion()
	if current == "" {
		return ErrNoCurrentVersion
	}

	return m.ExecVersion(ctx, current, program, args)
}

func (m *Manager) ExecVersion(
	ctx context.Context,
	version string,
	program string,
	args []string,
) error {
	ver, err := m.Get(ctx, version)
	if err != nil {
		return err
	}

	bin, err := ver.FindBin(program)
	if err != nil {
		return err
	}

	if ctx.Err() != nil {
		return ctx.Err()
	}

	execArgs := append([]string{bin}, args...)
	execEnv := os.Environ()

	// Prepend selected version's bin directory to PATH.
	for i := 0; i < len(execEnv); i++ {
		if strings.HasPrefix(execEnv[i], "PATH=") {
			execEnv[i] = "PATH=" + ver.BinDir + ":" + execEnv[i][5:]
		}
	}

	log.Debug().
		Str("bin", bin).
		Str("extra_path", ver.BinDir).
		Strs("args", args).
		Msg("executing")

	return syscall.Exec(bin, execArgs, execEnv)
}

func (m *Manager) FindBin(
	ctx context.Context,
	name string,
) ([]*Version, error) {
	versions, err := m.List(ctx)
	if err != nil {
		return nil, err
	}

	var availableIn []*Version
	for _, ver := range versions {
		if _, err := ver.FindBin(name); err == nil {
			availableIn = append(availableIn, ver)
		}
	}

	return availableIn, nil
}
