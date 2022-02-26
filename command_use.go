package main

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

func useCommand() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:               "use <version>",
		Short:             "Activate specified version.",
		Aliases:           []string{"activate", "switch"},
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: useValidArgs,
		RunE:              useRunE,
	}

	return cmd, nil
}

func useRunE(_ *cobra.Command, args []string) error {
	conf, err := getConfig()
	if err != nil {
		return err
	}

	version, err := newEmacsVersion(conf, args[0])
	if err != nil {
		return err
	}

	var shim bytes.Buffer
	err = useShimTemplate.Execute(&shim, conf)
	if err != nil {
		return err
	}

	err = os.MkdirAll(conf.Path.Shims, 0o755)
	if err != nil {
		return err
	}

	for _, bin := range version.Binaries {
		shimFile := filepath.Join(conf.Path.Shims, bin)
		err = os.WriteFile(shimFile, shim.Bytes(), 0o755)
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
	}

	currentFile := filepath.Join(conf.Path.Root, currentFileName)
	err = os.WriteFile(currentFile, []byte(version.Version), 0o644)
	if err != nil {
		return err
	}

	return nil
}

func useValidArgs(
	_ *cobra.Command,
	args []string,
	toComplete string,
) ([]string, cobra.ShellCompDirective) {
	var r []string

	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	conf, err := getConfig()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	versions, err := newEmacsVersions(conf)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	for _, ver := range versions {
		if toComplete == "" ||
			strings.HasPrefix(ver.Version, toComplete) {
			r = append(r, ver.Version)
		}
	}

	return r, cobra.ShellCompDirectiveNoFileComp
}

var useShimTemplate = template.Must(template.New("other").Parse(
	`#!/usr/bin/env bash
set -e
[ -n "$EVM_DEBUG" ] && set -x

program="${0##*/}"
export EVM_ROOT="{{.Path.Root}}"
exec "{{.Path.Binary}}" exec "$program" "$@"
`))
